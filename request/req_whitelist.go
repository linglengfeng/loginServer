package request

import (
	"errors"
	"fmt"
	"loginServer/config"
	"loginServer/src/db"
	"loginServer/src/log"
	"net"
	"strings"
)

// InitWhitelistFromDB 从数据库初始化白名单到缓存
func InitWhitelistFromDB() {
	// 从数据库加载所有白名单
	whitelists, err := db.LoadAllWhitelists()
	if err != nil {
		// 数据库加载失败，尝试从配置文件加载（兼容旧配置）
		InitWhitelistFromConfig()
		return
	}

	// 构建数据库中的白名单映射表：map[group][ip] = true，用于快速查找
	dbWhitelistMap := make(map[string]map[string]bool)
	groupMap := make(map[string][]string)
	for _, wl := range whitelists {
		groupLower := strings.ToLower(wl.APIGroup)
		ip := wl.IP

		// 初始化分组映射
		if dbWhitelistMap[groupLower] == nil {
			dbWhitelistMap[groupLower] = make(map[string]bool)
		}
		dbWhitelistMap[groupLower][ip] = true

		// 按分组组织数据用于缓存
		groupMap[groupLower] = append(groupMap[groupLower], ip)
	}

	// 检查配置文件，将配置中有但数据库中没有的IP添加到数据库
	allSettings := config.Config.AllSettings()
	if whitelistRaw, exists := allSettings["ip_whitelist"]; exists {
		if whitelistMap, ok := whitelistRaw.(map[string]interface{}); ok {
			for group, value := range whitelistMap {
				groupLower := strings.ToLower(group)

				// 解析配置中的IP列表
				var configIPs []string
				if ips, ok := value.([]interface{}); ok {
					for _, ip := range ips {
						if ipStr, ok := ip.(string); ok {
							ipStr = strings.TrimSpace(ipStr)
							if ipStr != "" {
								configIPs = append(configIPs, ipStr)
							}
						}
					}
				}

				// 检查配置中的每个IP是否在数据库中
				for _, configIP := range configIPs {
					// 如果该分组在数据库中不存在，或者该IP在该分组中不存在，则添加到数据库
					if dbWhitelistMap[groupLower] == nil || !dbWhitelistMap[groupLower][configIP] {
						// 添加到数据库（AddWhitelistIP 会自动处理已存在的情况）
						if err := db.AddWhitelistIP(groupLower, configIP); err == nil {
							// 添加成功，更新本地映射和缓存
							if dbWhitelistMap[groupLower] == nil {
								dbWhitelistMap[groupLower] = make(map[string]bool)
							}
							dbWhitelistMap[groupLower][configIP] = true
							groupMap[groupLower] = append(groupMap[groupLower], configIP)
						} else {
							// 记录错误日志，但不阻止启动
							log.Warn("从配置文件同步IP到数据库失败: 分组=%s, IP=%s, 错误=%v", groupLower, configIP, err)
						}
					}
				}
			}
		}
	}

	// 写入缓存
	for group, ips := range groupMap {
		setWhitelistToCache(group, ips)
	}
}

// InitWhitelistFromConfig 从配置文件初始化白名单（兼容旧配置，仅在数据库加载失败时使用）
func InitWhitelistFromConfig() {
	// 使用 AllSettings() 获取所有配置（保持原始键名大小写）
	allSettings := config.Config.AllSettings()
	whitelistRaw, exists := allSettings["ip_whitelist"]
	if !exists {
		return
	}

	whitelistMap, ok := whitelistRaw.(map[string]interface{})
	if !ok {
		return
	}

	for group, value := range whitelistMap {
		// 统一转换为小写存储，保证代码一致性（配置文件可以保持原始大小写）
		groupLower := strings.ToLower(group)

		var ipList []string

		// 解析 IP 列表
		if ips, ok := value.([]interface{}); ok {
			ipList = make([]string, 0, len(ips))
			for _, ip := range ips {
				if ipStr, ok := ip.(string); ok {
					ipStr = strings.TrimSpace(ipStr)
					if ipStr != "" {
						ipList = append(ipList, ipStr)
					}
				}
			}
		} else {
			ipList = []string{} // 设置为空数组
		}

		// 使用封装函数存储到缓存
		setWhitelistToCache(groupLower, ipList)
	}
}

// GetAllowedIPsByGroup 根据API分组获取IP白名单（线程安全）
// 返回值和含义：
//   - nil: 配置不存在，表示不限制IP（允许所有访问）
//   - []: 配置存在但为空列表，表示不允许任何IP访问
//   - [ip1, ip2, ...]: 配置了白名单IP列表
func GetAllowedIPsByGroup(apiGroup string) []string {
	// 统一转换为小写查找，保证代码一致性
	apiGroup = strings.ToLower(apiGroup)
	return getWhitelistFromCache(apiGroup)
}

// SetWhitelist 设置指定分组的IP白名单（完全替换）
func SetWhitelist(apiGroup string, ips []string) error {
	// 统一转换为小写，保证代码一致性
	apiGroup = strings.ToLower(apiGroup)
	// 验证IP格式
	validIPs := make([]string, 0, len(ips))
	for _, ip := range ips {
		ip = strings.TrimSpace(ip)
		if ip == "" {
			continue
		}

		// 验证IP格式（支持CIDR）
		if strings.Contains(ip, "/") {
			_, _, err := net.ParseCIDR(ip)
			if err != nil {
				return err
			}
		} else {
			if net.ParseIP(ip) == nil {
				return fmt.Errorf("无效的IP地址: %s", ip)
			}
		}
		validIPs = append(validIPs, ip)
	}

	// 写入数据库
	if err := db.SetWhitelist(apiGroup, validIPs); err != nil {
		return fmt.Errorf("保存白名单到数据库失败: %w", err)
	}

	// 更新缓存
	if len(validIPs) == 0 {
		setWhitelistToCache(apiGroup, []string{})
	} else {
		setWhitelistToCache(apiGroup, validIPs)
	}

	return nil
}

// AddIP 向指定分组添加IP（如果已存在则忽略）
func AddIP(apiGroup string, ip string) error {
	// 统一转换为小写，保证代码一致性
	apiGroup = strings.ToLower(apiGroup)
	ip = strings.TrimSpace(ip)
	if ip == "" {
		return errors.New("IP地址不能为空")
	}

	// 验证IP格式
	if strings.Contains(ip, "/") {
		_, _, err := net.ParseCIDR(ip)
		if err != nil {
			return err
		}
	} else {
		if net.ParseIP(ip) == nil {
			return fmt.Errorf("无效的IP地址: %s", ip)
		}
	}

	// 写入数据库
	if err := db.AddWhitelistIP(apiGroup, ip); err != nil {
		return fmt.Errorf("添加IP到数据库失败: %w", err)
	}

	// 从数据库重新加载该分组的白名单，确保缓存与数据库一致
	dbWhitelists, err := db.LoadWhitelist(apiGroup)
	if err != nil {
		// 如果重新加载失败，使用缓存更新逻辑（降级处理）
		ips := getWhitelistFromCache(apiGroup)
		if ips == nil {
			ips = []string{}
		}

		// 检查是否已存在（避免重复添加到缓存）
		exists := false
		for _, existingIP := range ips {
			if existingIP == ip {
				exists = true
				break
			}
		}

		if !exists {
			ips = append(ips, ip)
			setWhitelistToCache(apiGroup, ips)
		}
		return nil
	}

	// 将数据库中的数据转换为字符串数组并更新缓存
	ips := make([]string, 0, len(dbWhitelists))
	for _, wl := range dbWhitelists {
		ips = append(ips, wl.IP)
	}
	setWhitelistToCache(apiGroup, ips)

	return nil
}

// RemoveIP 从指定分组删除IP
func RemoveIP(apiGroup string, ip string) error {
	// 统一转换为小写，保证代码一致性
	apiGroup = strings.ToLower(apiGroup)
	ip = strings.TrimSpace(ip)
	if ip == "" {
		return errors.New("IP地址不能为空")
	}

	// 从数据库删除
	if err := db.RemoveWhitelistIP(apiGroup, ip); err != nil {
		return fmt.Errorf("从数据库删除IP失败: %w", err)
	}

	// 从数据库重新加载该分组的白名单，确保缓存与数据库一致
	dbWhitelists, err := db.LoadWhitelist(apiGroup)
	if err != nil {
		// 如果重新加载失败，使用缓存更新逻辑（降级处理）
		ips := getWhitelistFromCache(apiGroup)
		if ips == nil {
			return nil // 分组不存在，无需删除
		}

		// 查找并删除
		newIPs := make([]string, 0, len(ips))
		for _, existingIP := range ips {
			if existingIP != ip {
				newIPs = append(newIPs, existingIP)
			}
		}

		setWhitelistToCache(apiGroup, newIPs)
		return nil
	}

	// 将数据库中的数据转换为字符串数组并更新缓存
	ips := make([]string, 0, len(dbWhitelists))
	for _, wl := range dbWhitelists {
		ips = append(ips, wl.IP)
	}
	setWhitelistToCache(apiGroup, ips)

	return nil
}

// GetAllGroups 获取所有分组名称
func GetAllGroups() []string {
	items := getAllWhitelistItems()
	groups := make([]string, 0, len(items))
	prefix := CacheKeyWhitelist + "_"

	for key := range items {
		group := strings.TrimPrefix(key, prefix)
		groups = append(groups, group)
	}
	return groups
}

// GetWhitelist 获取指定分组的白名单（用于查询）
func GetWhitelist(apiGroup string) []string {
	return GetAllowedIPsByGroup(apiGroup)
}

// GetAllWhitelists 获取所有分组的白名单（用于查询）
func GetAllWhitelists() map[string][]string {
	items := getAllWhitelistItems()
	result := make(map[string][]string, len(items))
	prefix := CacheKeyWhitelist + "_"

	for key, item := range items {
		group := strings.TrimPrefix(key, prefix)
		if ips, ok := item.Object.([]string); ok {
			result[group] = make([]string, len(ips))
			copy(result[group], ips)
		}
	}
	return result
}
