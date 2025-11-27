package utils

import (
	"fmt"
	"math/rand"
	"time"

	"log-project/models"

	"github.com/google/uuid"
)

func GenerateSampleContent(size string) models.Content {
	// rand.Seed is automatically initialized in Go 1.20+

	switch size {
	case "small":
		return generateSmallContent()
	case "medium":
		return generateMediumContent()
	case "large":
		return generateLargeContent()
	default:
		return generateSmallContent()
	}
}

func generateSmallContent() models.Content {
	return models.Content{
		"event_id":    uuid.New().String(),
		"session_id":  fmt.Sprintf("sess_%d", rand.Intn(100000)),
		"ip_address":  fmt.Sprintf("192.168.%d.%d", rand.Intn(255), rand.Intn(255)),
		"user_agent":  getRandomUserAgent(),
		"timestamp":   time.Now().Unix(),
		"action_type": getRandomAction(),
		"status":      getRandomStatus(),
		"duration":    rand.Intn(5000) + 100,
		"device_id":   fmt.Sprintf("device_%d", rand.Intn(10000)),
	}
}

func generateMediumContent() models.Content {
	content := generateSmallContent()

	// Add more fields
	additionalFields := models.Content{
		"request_id":      uuid.New().String(),
		"correlation_id":  fmt.Sprintf("corr_%d", rand.Intn(1000000)),
		"source_ip":       fmt.Sprintf("10.0.%d.%d", rand.Intn(255), rand.Intn(255)),
		"destination_ip":  fmt.Sprintf("172.16.%d.%d", rand.Intn(255), rand.Intn(255)),
		"protocol":        getRandomProtocol(),
		"port":            rand.Intn(65535),
		"bytes_sent":      rand.Intn(1000000),
		"bytes_received":  rand.Intn(1000000),
		"latency":         rand.Intn(1000),
		"error_code":      getRandomErrorCode(),
		"retry_count":     rand.Intn(5),
		"cache_hit":       rand.Intn(2) == 1,
		"compression":     rand.Intn(2) == 1,
		"encrypted":       rand.Intn(2) == 1,
		"region":          getRandomRegion(),
		"datacenter":      fmt.Sprintf("dc-%d", rand.Intn(10)+1),
		"service_version": fmt.Sprintf("v%d.%d.%d", rand.Intn(5)+1, rand.Intn(10), rand.Intn(20)),
		"build_number":    rand.Intn(10000),
		"environment":     getRandomEnvironment(),
		"tenant_id":       uuid.New().String(),
		"org_id":          fmt.Sprintf("org_%d", rand.Intn(1000)),
		"team_id":         fmt.Sprintf("team_%d", rand.Intn(100)),
		"project_id":      fmt.Sprintf("proj_%d", rand.Intn(50)),
		"feature_flag":    getRandomFeatureFlag(),
		"ab_test":         fmt.Sprintf("test_%d", rand.Intn(100)),
		"experiment_id":   uuid.New().String(),
		"segment":         getRandomSegment(),
		"cohort":          fmt.Sprintf("cohort_%d", rand.Intn(10)+1),
		"tier":            getRandomTier(),
		"plan":            getRandomPlan(),
		"quota":           rand.Intn(10000),
		"usage":           rand.Intn(1000),
		"limit":           rand.Intn(5000),
		"remaining":       rand.Intn(1000),
		"renewal_date":    time.Now().AddDate(0, rand.Intn(12), 0).Format("2006-01-02"),
		"last_login":      time.Now().Add(-time.Duration(rand.Intn(86400)) * time.Second).Format(time.RFC3339),
		"first_login":     time.Now().Add(-time.Duration(rand.Intn(86400*30)) * time.Second).Format(time.RFC3339),
		"session_count":   rand.Intn(100),
		"total_sessions":  rand.Intn(1000),
		"country":         getRandomCountry(),
		"city":            getRandomCity(),
		"timezone":        getRandomTimezone(),
		"language":        getRandomLanguage(),
		"currency":        getRandomCurrency(),
		"description":     generateJapaneseString(rand.Intn(100) + 20),
		"notes":           generateJapaneseString(rand.Intn(50) + 10),
	}

	for k, v := range additionalFields {
		content[k] = v
	}

	return content
}

func generateLargeContent() models.Content {
	content := generateMediumContent()

	// Add many more fields for large content
	for i := 0; i < 500; i++ {
		content[fmt.Sprintf("field_%d", i)] = fmt.Sprintf("value_%d_%s", rand.Intn(100000), generateJapaneseString(5))
	}

	// Add Japanese content fields
	for i := 0; i < 200; i++ {
		content[fmt.Sprintf("japanese_field_%d", i)] = generateJapaneseString(rand.Intn(100) + 20)
	}

	// Add nested objects
	for i := 0; i < 100; i++ {
		nested := make(models.Content)
		for j := 0; j < 10; j++ {
			nested[fmt.Sprintf("nested_field_%d", j)] = fmt.Sprintf("nested_value_%d_%s", rand.Intn(1000), generateJapaneseString(5))
		}
		content[fmt.Sprintf("nested_obj_%d", i)] = nested
	}

	// Add arrays
	for i := 0; i < 50; i++ {
		arr := make([]string, rand.Intn(20)+1)
		for j := range arr {
			arr[j] = fmt.Sprintf("array_item_%d_%d_%s", i, j, generateJapaneseString(5))
		}
		content[fmt.Sprintf("array_field_%d", i)] = arr
	}

	return content
}

func getRandomUserAgent() string {
	userAgents := []string{
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36",
		"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36",
		"Mozilla/5.0 (iPhone; CPU iPhone OS 14_7_1 like Mac OS X)",
		"Mozilla/5.0 (Android 11; Mobile; rv:68.0) Gecko/68.0 Firefox/88.0",
	}
	return userAgents[rand.Intn(len(userAgents))]
}

func getRandomAction() string {
	actions := []string{
		"login", "logout", "view", "click", "purchase", "search", "filter",
		"create", "update", "delete", "download", "upload", "share", "comment",
		"like", "dislike", "subscribe", "unsubscribe", "follow", "unfollow",
	}
	return actions[rand.Intn(len(actions))]
}

func getRandomStatus() string {
	statuses := []string{
		"success", "failure", "pending", "timeout", "error", "warning", "info",
	}
	return statuses[rand.Intn(len(statuses))]
}

func getRandomProtocol() string {
	protocols := []string{"HTTP", "HTTPS", "WebSocket", "gRPC", "TCP", "UDP"}
	return protocols[rand.Intn(len(protocols))]
}

func getRandomErrorCode() string {
	codes := []string{"200", "201", "400", "401", "403", "404", "500", "502", "503", "504"}
	return codes[rand.Intn(len(codes))]
}

func getRandomRegion() string {
	regions := []string{"us-east-1", "us-west-2", "eu-west-1", "ap-southeast-1", "ap-northeast-1"}
	return regions[rand.Intn(len(regions))]
}

func getRandomEnvironment() string {
	envs := []string{"production", "staging", "development", "testing"}
	return envs[rand.Intn(len(envs))]
}

func getRandomFeatureFlag() string {
	flags := []string{"new_ui", "beta_search", "advanced_analytics", "real_time_sync", "auto_backup"}
	return flags[rand.Intn(len(flags))]
}

func getRandomSegment() string {
	segments := []string{"premium", "basic", "trial", "enterprise", "free"}
	return segments[rand.Intn(len(segments))]
}

func getRandomTier() string {
	tiers := []string{"bronze", "silver", "gold", "platinum", "diamond"}
	return tiers[rand.Intn(len(tiers))]
}

func getRandomPlan() string {
	plans := []string{"starter", "professional", "business", "enterprise", "custom"}
	return plans[rand.Intn(len(plans))]
}

func getRandomCountry() string {
	countries := []string{"US", "UK", "CA", "AU", "DE", "FR", "JP", "SG", "IN", "BR"}
	return countries[rand.Intn(len(countries))]
}

func getRandomCity() string {
	cities := []string{"New York", "London", "Toronto", "Sydney", "Berlin", "Paris", "Tokyo", "Singapore", "Mumbai", "São Paulo"}
	return cities[rand.Intn(len(cities))]
}

func getRandomTimezone() string {
	timezones := []string{"UTC", "America/New_York", "Europe/London", "Asia/Tokyo", "Australia/Sydney"}
	return timezones[rand.Intn(len(timezones))]
}

func getRandomLanguage() string {
	languages := []string{"en", "es", "fr", "de", "ja", "zh", "pt", "ru", "ar", "hi"}
	return languages[rand.Intn(len(languages))]
}

func getRandomCurrency() string {
	currencies := []string{"USD", "EUR", "GBP", "JPY", "CAD", "AUD", "CHF", "CNY", "INR", "BRL"}
	return currencies[rand.Intn(len(currencies))]
}

func generateJapaneseString(length int) string {
	// Hiragana: 0x3040 - 0x309F
	// Katakana: 0x30A0 - 0x30FF
	// Kanji: 0x4E00 - 0x9FAF

	runes := make([]rune, length)
	for i := 0; i < length; i++ {
		type_ := rand.Intn(3)
		switch type_ {
		case 0: // Hiragana
			runes[i] = rune(0x3040 + rand.Intn(0x309F-0x3040+1))
		case 1: // Katakana
			runes[i] = rune(0x30A0 + rand.Intn(0x30FF-0x30A0+1))
		case 2: // Kanji (subset for simplicity)
			runes[i] = rune(0x4E00 + rand.Intn(0x1000))
		}
	}
	return string(runes)
}
