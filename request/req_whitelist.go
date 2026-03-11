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
func InitWhitelistFromDB() error {
	// 1. 从数据库加载所有白名单
	whitelists, err := db.LoadAllWhitelists()
	if err != nil {
		return fmt.Errorf("从数据库加载白名单失败: %w", err)
	}

	// 构建数据库中的白名单映射表：map[group][ip] = true，用于快速查找
	dbWhitelistMap := make(map[string]map[string]bool)
	groupMap := make(map[string][]string)
	groupIPSet := make(map[string]map[string]bool) // 用于去重检查
	for _, wl := range whitelists {
		groupLower := strings.ToLower(wl.APIGroup)
		ip := wl.IP

		// 初始化分组映射
		if dbWhitelistMap[groupLower] == nil {
			dbWhitelistMap[groupLower] = make(map[string]bool)
		}
		dbWhitelistMap[groupLower][ip] = true

		// 按分组组织数据用于缓存（去重）
		if groupIPSet[groupLower] == nil {
			groupIPSet[groupLower] = make(map[string]bool)
		}
		if !groupIPSet[groupLower][ip] {
			groupIPSet[groupLower][ip] = true
			groupMap[groupLower] = append(groupMap[groupLower], ip)
		}
	}

	// 2. 对比配置，配置有的但数据库中没有的，加入到数据库
	allSettings := config.Config.AllSettings()
	if whitelistRaw, exists := allSettings["ip_whitelist"]; exists {
		if whitelistMap, ok := whitelistRaw.(map[string]any); ok {
			for group, value := range whitelistMap {
				groupLower := strings.ToLower(group)

				// 确保该分组在 groupMap 和 groupIPSet 中存在（即使IP列表为空）
				if groupMap[groupLower] == nil {
					groupMap[groupLower] = []string{}
				}
				if groupIPSet[groupLower] == nil {
					groupIPSet[groupLower] = make(map[string]bool)
				}

				// 解析配置中的IP列表
				var configIPs []string
				has127001 := false
				if ips, ok := value.([]any); ok {
					for _, ip := range ips {
						if ipStr, ok := ip.(string); ok {
							ipStr = strings.TrimSpace(ipStr)
							if ipStr != "" {
								configIPs = append(configIPs, ipStr)
								if ipStr == "127.0.0.1" {
									has127001 = true
								}
							}
						}
					}
				}

				// 检查配置中的每个IP是否在数据库中，如果不在则添加到数据库
				for _, configIP := range configIPs {
					// 如果该分组在数据库中不存在，或者该IP在该分组中不存在，则添加到数据库
					if dbWhitelistMap[groupLower] == nil || !dbWhitelistMap[groupLower][configIP] {
						// 先修改数据库
						if err := db.AddWhitelistIP(groupLower, configIP); err == nil {
							// 添加成功，更新本地映射
							if dbWhitelistMap[groupLower] == nil {
								dbWhitelistMap[groupLower] = make(map[string]bool)
							}
							dbWhitelistMap[groupLower][configIP] = true
							// 使用 groupIPSet 进行去重检查
							if groupIPSet[groupLower] == nil {
								groupIPSet[groupLower] = make(map[string]bool)
							}
							if !groupIPSet[groupLower][configIP] {
								groupIPSet[groupLower][configIP] = true
								groupMap[groupLower] = append(groupMap[groupLower], configIP)
							}
						} else {
							// 记录错误日志，但不阻止启动
							log.Warn("从配置文件同步IP到数据库失败: 分组=%s, IP=%s, 错误=%v", groupLower, configIP, err)
						}
					} else {
						// IP已在数据库中，确保在 groupMap 中（可能第一步已加载）
						// 使用 groupIPSet 进行去重检查
						if groupIPSet[groupLower] == nil {
							groupIPSet[groupLower] = make(map[string]bool)
						}
						if !groupIPSet[groupLower][configIP] {
							groupIPSet[groupLower][configIP] = true
							groupMap[groupLower] = append(groupMap[groupLower], configIP)
						}
					}
				}

				// 如果配置了 127.0.0.1，自动添加服务器绑定的 IP 和 IPv6 回环地址
				if has127001 {
					// 获取服务器绑定的 IP
					serverIP := config.Config.GetString("gin.ip")
					if serverIP == "" || serverIP == "127.0.0.1" || serverIP == "localhost" {
						// 如果配置是 127.0.0.1 或 localhost，获取真实局域网 IP
						realIP, err := getLocalIPv4()
						if err == nil {
							serverIP = realIP
						}
					}

					// 如果服务器绑定的 IP 不是 127.0.0.1，且不在数据库中，则添加到数据库
					if serverIP != "" && serverIP != "127.0.0.1" && serverIP != "localhost" {
						if dbWhitelistMap[groupLower] == nil || !dbWhitelistMap[groupLower][serverIP] {
							// 先修改数据库
							if err := db.AddWhitelistIP(groupLower, serverIP); err == nil {
								// 添加成功，更新本地映射
								if dbWhitelistMap[groupLower] == nil {
									dbWhitelistMap[groupLower] = make(map[string]bool)
								}
								dbWhitelistMap[groupLower][serverIP] = true
								// 使用 groupIPSet 进行去重检查
								if groupIPSet[groupLower] == nil {
									groupIPSet[groupLower] = make(map[string]bool)
								}
								if !groupIPSet[groupLower][serverIP] {
									groupIPSet[groupLower][serverIP] = true
									groupMap[groupLower] = append(groupMap[groupLower], serverIP)
								}
							} else {
								// 记录错误日志，但不阻止启动
								log.Warn("自动添加服务器绑定IP到数据库失败: 分组=%s, IP=%s, 错误=%v", groupLower, serverIP, err)
							}
						} else {
							// IP已在数据库中，确保在 groupMap 中
							// 使用 groupIPSet 进行去重检查
							if groupIPSet[groupLower] == nil {
								groupIPSet[groupLower] = make(map[string]bool)
							}
							if !groupIPSet[groupLower][serverIP] {
								groupIPSet[groupLower][serverIP] = true
								groupMap[groupLower] = append(groupMap[groupLower], serverIP)
							}
						}
					}

					// 自动添加 IPv6 回环地址到缓存（不写入数据库）
					hasIPv6Loopback := false
					for _, ip := range groupMap[groupLower] {
						if ip == "::1" {
							hasIPv6Loopback = true
							break
						}
					}
					// 如果缓存中没有 ::1，且数据库中也没有，则添加到缓存
					if !hasIPv6Loopback && (dbWhitelistMap[groupLower] == nil || !dbWhitelistMap[groupLower]["::1"]) {
						if groupIPSet[groupLower] == nil {
							groupIPSet[groupLower] = make(map[string]bool)
						}
						if !groupIPSet[groupLower]["::1"] {
							groupIPSet[groupLower]["::1"] = true
							groupMap[groupLower] = append(groupMap[groupLower], "::1")
						}
					}
				}
			}
		}
	}

	// 3. 写入缓存
	for group, ips := range groupMap {
		db.SetWhitelistToCache(group, ips)
	}

	return nil
}

// GetAllowedIPsByGroup 根据API分组获取IP白名单（线程安全）
// 返回值和含义：
//   - nil: 配置不存在，表示不限制IP（允许所有访问）
//   - []: 配置存在但为空列表，表示不允许任何IP访问
//   - [ip1, ip2, ...]: 配置了白名单IP列表
func GetAllowedIPsByGroup(apiGroup string) []string {
	apiGroup = strings.ToLower(apiGroup)
	return db.GetWhitelistFromCache(apiGroup)
}

// SetWhitelist 设置指定分组的IP白名单（完全替换）
func SetWhitelist(apiGroup string, ips []string) error {
	apiGroup = strings.ToLower(apiGroup)
	validIPs := make([]string, 0, len(ips))
	ipSet := make(map[string]bool)
	for _, ip := range ips {
		ip = strings.TrimSpace(ip)
		if ip == "" {
			continue
		}
		if ipSet[ip] {
			continue
		}
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
		ipSet[ip] = true
		validIPs = append(validIPs, ip)
	}

	if err := db.SetWhitelist(apiGroup, validIPs); err != nil {
		return fmt.Errorf("保存白名单到数据库失败: %w", err)
	}

	if len(validIPs) == 0 {
		db.SetWhitelistToCache(apiGroup, []string{})
	} else {
		db.SetWhitelistToCache(apiGroup, validIPs)
	}

	return nil
}

// AddIP 向指定分组添加IP（如果已存在则忽略）
func AddIP(apiGroup string, ip string) error {
	apiGroup = strings.ToLower(apiGroup)
	ip = strings.TrimSpace(ip)
	if ip == "" {
		return errors.New("IP地址不能为空")
	}

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

	if err := db.AddWhitelistIP(apiGroup, ip); err != nil {
		return fmt.Errorf("添加IP到数据库失败: %w", err)
	}

	dbWhitelists, err := db.LoadWhitelist(apiGroup)
	if err != nil {
		log.Warn("AddIP: 重新加载白名单失败，使用降级处理: 分组=%s, 错误=%v", apiGroup, err)
		ips := db.GetWhitelistFromCache(apiGroup)
		if ips == nil {
			ips = []string{}
		}
		exists := false
		for _, existingIP := range ips {
			if existingIP == ip {
				exists = true
				break
			}
		}
		if !exists {
			ips = append(ips, ip)
			db.SetWhitelistToCache(apiGroup, ips)
		}
		return nil
	}

	ips := make([]string, 0, len(dbWhitelists))
	for _, wl := range dbWhitelists {
		ips = append(ips, wl.IP)
	}
	db.SetWhitelistToCache(apiGroup, ips)

	return nil
}

// RemoveIP 从指定分组删除IP
func RemoveIP(apiGroup string, ip string) error {
	apiGroup = strings.ToLower(apiGroup)
	ip = strings.TrimSpace(ip)
	if ip == "" {
		return errors.New("IP地址不能为空")
	}

	if err := db.RemoveWhitelistIP(apiGroup, ip); err != nil {
		return fmt.Errorf("从数据库删除IP失败: %w", err)
	}

	dbWhitelists, err := db.LoadWhitelist(apiGroup)
	if err != nil {
		log.Warn("RemoveIP: 重新加载白名单失败，使用降级处理: 分组=%s, 错误=%v", apiGroup, err)
		ips := db.GetWhitelistFromCache(apiGroup)
		if ips == nil {
			return nil
		}
		newIPs := make([]string, 0, len(ips))
		for _, existingIP := range ips {
			if existingIP != ip {
				newIPs = append(newIPs, existingIP)
			}
		}
		db.SetWhitelistToCache(apiGroup, newIPs)
		return nil
	}

	ips := make([]string, 0, len(dbWhitelists))
	for _, wl := range dbWhitelists {
		ips = append(ips, wl.IP)
	}
	db.SetWhitelistToCache(apiGroup, ips)

	return nil
}

// GetAllGroups 获取所有分组名称
func GetAllGroups() []string {
	items := db.GetAllWhitelistItems()
	groups := make([]string, 0, len(items))
	prefix := db.IPWhitelistCacheKeyPrefix

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
	items := db.GetAllWhitelistItems()
	result := make(map[string][]string, len(items))
	prefix := db.IPWhitelistCacheKeyPrefix

	for key, ips := range items {
		group := strings.TrimPrefix(key, prefix)
		result[group] = make([]string, len(ips))
		copy(result[group], ips)
	}
	return result
}

// getLocalIPv4 获取本机首个非回环的 IPv4 地址
func getLocalIPv4() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}

	for _, address := range addrs {
		// 检查 ip 地址判断是否回环地址
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			// 必须是 IPv4
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String(), nil
			}
		}
	}

	return "", errors.New("cannot find local IP address")
}
