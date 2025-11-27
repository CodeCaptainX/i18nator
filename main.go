package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"
)

// Language configuration
var languages = map[string]string{
	"en.json": "en",
	"km.json": "km",
	"zh.json": "zh-CN",
}

var basePath string

func main() {
	var rootCmd = &cobra.Command{
		Use:   "i18nator",
		Short: "A CLI tool to manage i18n JSON files",
		Long:  "i18nator helps you manage internationalization files with automatic translation support",
	}

	// Add command
	var addCmd = &cobra.Command{
		Use:   "add [key] [value]",
		Short: "Add a new i18n key with automatic translation",
		Long:  "Add a new key-value pair to all language files with automatic translation",
		Args:  cobra.ExactArgs(2),
		Run:   runAdd,
	}

	// List command
	var listCmd = &cobra.Command{
		Use:   "list",
		Short: "List all i18n keys",
		Long:  "Display all i18n keys from the English language file",
		Run:   runList,
	}

	// Update command
	var updateCmd = &cobra.Command{
		Use:   "update [key] [value]",
		Short: "Update an existing i18n key",
		Long:  "Update an existing key-value pair in all language files with automatic translation",
		Args:  cobra.ExactArgs(2),
		Run:   runUpdate,
	}

	// Remove command
	var removeCmd = &cobra.Command{
		Use:   "remove [key]",
		Short: "Remove an i18n key from all languages",
		Long:  "Remove a key-value pair from all language files",
		Args:  cobra.ExactArgs(1),
		Run:   runRemove,
	}

	// Add commands to root
	rootCmd.AddCommand(addCmd, listCmd, updateCmd, removeCmd)

	// Execute
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// Initialize base path
func initBasePath() error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %v", err)
	}
	basePath = filepath.Join(cwd, "pkg", "translates", "localize", "i18n")

	// Create directory if it doesn't exist
	if err := os.MkdirAll(basePath, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create directory: %v", err)
	}

	fmt.Printf("📂 Using base path: %s\n", basePath)
	return nil
}

// Add command handler
func runAdd(cmd *cobra.Command, args []string) {
	if err := initBasePath(); err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		return
	}

	key := args[0]
	value := args[1]

	// Check if key already exists
	enPath := filepath.Join(basePath, "en.json")
	enData := loadJSON(enPath)
	if _, exists := enData[key]; exists {
		fmt.Printf("⚠️  Key '%s' already exists. Use 'update' command to modify it.\n", key)
		return
	}

	fmt.Println()
	for file, langCode := range languages {
		path := filepath.Join(basePath, file)
		processAddOrUpdate(path, key, value, file == "en.json", langCode)
	}

	fmt.Println("\n✨ i18n key added to all languages!")
}

// List command handler
func runList(cmd *cobra.Command, args []string) {
	if err := initBasePath(); err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		return
	}

	enPath := filepath.Join(basePath, "en.json")
	data := loadJSON(enPath)

	if len(data) == 0 {
		fmt.Println("📭 No i18n keys found")
		return
	}

	fmt.Printf("📋 Found %d i18n keys:\n\n", len(data))

	// Sort keys
	keys := make([]string, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Display
	for _, k := range keys {
		fmt.Printf("  %s: %s\n", k, data[k])
	}
}

// Update command handler
func runUpdate(cmd *cobra.Command, args []string) {
	if err := initBasePath(); err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		return
	}

	key := args[0]
	value := args[1]

	// Check if key exists
	enPath := filepath.Join(basePath, "en.json")
	enData := loadJSON(enPath)
	if _, exists := enData[key]; !exists {
		fmt.Printf("⚠️  Key '%s' does not exist. Use 'add' command to create it.\n", key)
		return
	}

	fmt.Println()
	for file, langCode := range languages {
		path := filepath.Join(basePath, file)
		processAddOrUpdate(path, key, value, file == "en.json", langCode)
	}

	fmt.Println("\n✨ i18n key updated in all languages!")
}

// Remove command handler
func runRemove(cmd *cobra.Command, args []string) {
	if err := initBasePath(); err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		return
	}

	key := args[0]

	// Check if key exists
	enPath := filepath.Join(basePath, "en.json")
	enData := loadJSON(enPath)
	if _, exists := enData[key]; !exists {
		fmt.Printf("⚠️  Key '%s' does not exist\n", key)
		return
	}

	fmt.Println()
	for file := range languages {
		path := filepath.Join(basePath, file)
		data := loadJSON(path)

		if _, exists := data[key]; exists {
			delete(data, key)
			saveJSON(path, data)
			fmt.Printf("✅ %s: Key removed\n", file)
		} else {
			fmt.Printf("⏭️  %s: Key not found\n", file)
		}
	}

	fmt.Println("\n✨ i18n key removed from all languages!")
}

// Process add or update operation
func processAddOrUpdate(path, key, value string, isEnglish bool, langCode string) {
	data := loadJSON(path)

	if isEnglish {
		data[key] = value
		saveJSON(path, data)
		fmt.Printf("✅ %s: %s\n", filepath.Base(path), value)
	} else {
		translated, err := googleTranslate(value, langCode)
		if err != nil {
			fmt.Printf("❌ %s: Translation failed (%v), using English\n", filepath.Base(path), err)
			translated = value
		}
		data[key] = translated
		saveJSON(path, data)
		fmt.Printf("✅ %s: %s\n", filepath.Base(path), translated)
	}
}

// Load JSON file
func loadJSON(path string) map[string]string {
	fileData, err := os.ReadFile(path)
	if err != nil {
		return map[string]string{}
	}

	var data map[string]string
	if err := json.Unmarshal(fileData, &data); err != nil {
		return map[string]string{}
	}

	if data == nil {
		data = map[string]string{}
	}
	return data
}

// Save JSON file with sorted keys
func saveJSON(path string, data map[string]string) error {
	keys := make([]string, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	sorted := make(map[string]string)
	for _, k := range keys {
		sorted[k] = data[k]
	}

	jsonBytes, err := json.MarshalIndent(sorted, "", "  ")
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), os.ModePerm); err != nil {
		return err
	}

	return os.WriteFile(path, jsonBytes, 0644)
}

// Google Translate using free API
func googleTranslate(text, targetLang string) (string, error) {
	baseURL := "https://translate.googleapis.com/translate_a/single"

	params := url.Values{}
	params.Add("client", "gtx")
	params.Add("sl", "en")
	params.Add("tl", targetLang)
	params.Add("dt", "t")
	params.Add("q", text)

	fullURL := baseURL + "?" + params.Encode()

	resp, err := http.Get(fullURL)
	if err != nil {
		return "", fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %v", err)
	}

	var result []interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %v", err)
	}

	if len(result) == 0 {
		return "", fmt.Errorf("empty response")
	}

	translations, ok := result[0].([]interface{})
	if !ok || len(translations) == 0 {
		return "", fmt.Errorf("invalid response format")
	}

	var translatedText strings.Builder
	for _, item := range translations {
		if arr, ok := item.([]interface{}); ok && len(arr) > 0 {
			if str, ok := arr[0].(string); ok {
				translatedText.WriteString(str)
			}
		}
	}

	return strings.TrimSpace(translatedText.String()), nil
}
