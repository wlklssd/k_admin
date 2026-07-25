package kadmin

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/GoAdminGroup/go-admin/plugins/admin/models"
	"github.com/gin-gonic/gin"
)

const systemConfigPathEnv = "KADMIN_SYSTEM_CONFIG_PATH"

type systemConfigItem struct {
	Key         string   `json:"key"`
	Value       string   `json:"value"`
	Label       string   `json:"label,omitempty"`
	Type        string   `json:"type,omitempty"`
	Options     []string `json:"options,omitempty"`
	Description string   `json:"description,omitempty"`
	Builtin     bool     `json:"builtin"`
	Future      bool     `json:"future,omitempty"`
}

type systemConfigPayload struct {
	Items []systemConfigItem `json:"items"`
}

type systemConfigOverview struct {
	FilePath string             `json:"filePath"`
	Items    []systemConfigItem `json:"items"`
}

type publicLoginConfig struct {
	CaptchaEnabled bool `json:"captchaEnabled"`
}

type systemConfigMeta struct {
	Key         string
	Label       string
	Type        string
	Default     string
	Options     []string
	Description string
	Future      bool
	Order       int
}

var systemConfigMetas = []systemConfigMeta{
	{
		Key:         "auth.default_username",
		Label:       "默认账号",
		Type:        "text",
		Default:     "admin",
		Description: "本地默认管理员账号",
		Order:       10,
	},
	{
		Key:         "auth.default_password",
		Label:       "默认密码",
		Type:        "password",
		Default:     "admin",
		Description: "本地默认管理员密码",
		Order:       20,
	},
	{
		Key:         "security.captcha_enabled",
		Label:       "验证码",
		Type:        "boolean",
		Default:     "false",
		Description: "登录验证码开关",
		Order:       30,
	},
	{
		Key:         "ui.theme_mode",
		Label:       "界面模式",
		Type:        "select",
		Default:     "auto",
		Options:     []string{"auto", "light", "dark"},
		Description: "跟随电脑主题 / 白天 / 黑夜模式",
		Order:       40,
	},
}

func registerSystemConfigRoutes(api *gin.RouterGroup, s *Store) {
	api.GET("/system/config/login", s.publicLoginConfig)
	api.GET("/system/config/login/", s.publicLoginConfig)

	group := api.Group("/system/config", s.requireAuth(), s.requireAdmin())
	group.GET("", s.systemConfig)
	group.GET("/", s.systemConfig)
	group.PUT("", s.updateSystemConfig)
	group.PUT("/", s.updateSystemConfig)
}

func (s *Store) publicLoginConfig(c *gin.Context) {
	values, err := s.readSystemConfig()
	if err != nil {
		fail(c, http.StatusInternalServerError, err.Error())
		return
	}
	success(c, publicLoginConfig{
		CaptchaEnabled: normalizeSystemConfigValue(
			systemConfigMeta{Type: "boolean"},
			values["security.captcha_enabled"],
		) == "true",
	})
}

func (s *Store) systemConfig(c *gin.Context) {
	values, err := s.readSystemConfig()
	if err != nil {
		fail(c, http.StatusInternalServerError, err.Error())
		return
	}
	success(c, systemConfigOverview{
		FilePath: systemConfigPath(),
		Items:    buildSystemConfigItems(values),
	})
}

func (s *Store) updateSystemConfig(c *gin.Context) {
	var req systemConfigPayload
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "invalid system config payload")
		return
	}

	values, err := normalizeSystemConfigPayload(req.Items)
	if err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	if err := s.writeSystemConfig(values); err != nil {
		fail(c, http.StatusInternalServerError, err.Error())
		return
	}

	success(c, systemConfigOverview{
		FilePath: systemConfigPath(),
		Items:    buildSystemConfigItems(values),
	})
}

func (s *Store) checkConfiguredDefaultAdmin(username, password string) (models.UserModel, bool) {
	values, err := s.readSystemConfig()
	if err != nil {
		return models.User(), false
	}

	configUsername := strings.TrimSpace(values["auth.default_username"])
	configPassword := values["auth.default_password"]
	if configUsername == "" || username != configUsername || password != configPassword {
		return models.User(), false
	}

	user := models.User().SetConn(s.conn).FindByUserName(configUsername)
	if user.IsEmpty() {
		return models.User(), false
	}
	return user.WithRoles().WithPermissions().WithMenus(), true
}

func (s *Store) readSystemConfig() (map[string]string, error) {
	s.configMu.Lock()
	defer s.configMu.Unlock()

	path := systemConfigPath()
	values := defaultSystemConfigValues()

	content, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			if err := ensureSystemConfigDir(path); err != nil {
				return nil, err
			}
			if err := writeSystemConfigFile(path, values); err != nil {
				return nil, err
			}
			return values, nil
		}
		return nil, err
	}
	if len(strings.TrimSpace(string(content))) == 0 {
		return values, nil
	}

	var fileValues map[string]string
	if err := json.Unmarshal(content, &fileValues); err != nil {
		return nil, err
	}
	for key, value := range fileValues {
		key = strings.TrimSpace(key)
		if key != "" {
			values[key] = value
		}
	}
	return values, nil
}

func (s *Store) writeSystemConfig(values map[string]string) error {
	s.configMu.Lock()
	defer s.configMu.Unlock()

	path := systemConfigPath()
	if err := ensureSystemConfigDir(path); err != nil {
		return err
	}
	return writeSystemConfigFile(path, values)
}

func systemConfigPath() string {
	if path := strings.TrimSpace(os.Getenv(systemConfigPathEnv)); path != "" {
		return filepath.Clean(path)
	}
	return filepath.Clean(filepath.Join("data", "system_config.json"))
}

func ensureSystemConfigDir(path string) error {
	dir := filepath.Dir(path)
	if dir == "." || dir == "" {
		return nil
	}
	return os.MkdirAll(dir, 0755)
}

func writeSystemConfigFile(path string, values map[string]string) error {
	content, err := json.MarshalIndent(values, "", "  ")
	if err != nil {
		return err
	}
	content = append(content, '\n')
	return os.WriteFile(path, content, 0644)
}

func defaultSystemConfigValues() map[string]string {
	values := make(map[string]string, len(systemConfigMetas))
	for _, meta := range systemConfigMetas {
		values[meta.Key] = meta.Default
	}
	return values
}

func normalizeSystemConfigPayload(items []systemConfigItem) (map[string]string, error) {
	values := defaultSystemConfigValues()
	seen := make(map[string]bool, len(items))

	for _, item := range items {
		key := strings.TrimSpace(item.Key)
		if key == "" {
			return nil, errors.New("config key is required")
		}
		if seen[key] {
			return nil, errors.New("duplicate config key: " + key)
		}
		seen[key] = true
		if !validSystemConfigKey(key) {
			return nil, errors.New("invalid config key: " + key)
		}

		value := item.Value
		if meta, ok := systemConfigMetaByKey(key); ok {
			value = normalizeSystemConfigValue(meta, value)
		}
		values[key] = value
	}
	return values, nil
}

func validSystemConfigKey(key string) bool {
	for _, r := range key {
		if r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' {
			continue
		}
		switch r {
		case '.', '_', '-', ':':
			continue
		default:
			return false
		}
	}
	return true
}

func normalizeSystemConfigValue(meta systemConfigMeta, value string) string {
	switch meta.Type {
	case "boolean":
		switch strings.ToLower(strings.TrimSpace(value)) {
		case "1", "true", "yes", "on":
			return "true"
		default:
			return "false"
		}
	case "select":
		value = strings.TrimSpace(value)
		for _, option := range meta.Options {
			if value == option {
				return value
			}
		}
		return meta.Default
	default:
		return strings.TrimSpace(value)
	}
}

func buildSystemConfigItems(values map[string]string) []systemConfigItem {
	items := make([]systemConfigItem, 0, len(values))
	added := make(map[string]bool, len(values))

	metas := append([]systemConfigMeta(nil), systemConfigMetas...)
	sort.SliceStable(metas, func(i, j int) bool {
		return metas[i].Order < metas[j].Order
	})
	for _, meta := range metas {
		value, ok := values[meta.Key]
		if !ok {
			value = meta.Default
		}
		items = append(items, systemConfigItem{
			Key:         meta.Key,
			Value:       normalizeSystemConfigValue(meta, value),
			Label:       meta.Label,
			Type:        meta.Type,
			Options:     meta.Options,
			Description: meta.Description,
			Builtin:     true,
			Future:      meta.Future,
		})
		added[meta.Key] = true
	}

	customKeys := make([]string, 0, len(values))
	for key := range values {
		if !added[key] {
			customKeys = append(customKeys, key)
		}
	}
	sort.Strings(customKeys)
	for _, key := range customKeys {
		items = append(items, systemConfigItem{
			Key:     key,
			Value:   values[key],
			Label:   key,
			Type:    "text",
			Builtin: false,
		})
	}

	return items
}

func systemConfigMetaByKey(key string) (systemConfigMeta, bool) {
	for _, meta := range systemConfigMetas {
		if meta.Key == key {
			return meta, true
		}
	}
	return systemConfigMeta{}, false
}
