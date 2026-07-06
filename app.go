package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

// App struct
type App struct {
	ctx context.Context
}

// Feature describes one toggleable modpack feature group.
type Feature struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Hashes      []string `json:"mods"`
}

// ModStatus describes one mod file inside a feature, for the expanded view.
type ModStatus struct {
	DisplayName string `json:"displayName"`
	Enabled     bool   `json:"enabled"`
	Found       bool   `json:"found"`
}

// FeatureStatus is what the frontend renders per card.
type FeatureStatus struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Enabled     bool        `json:"enabled"`
	Mods        []ModStatus `json:"mods"`
}

var features = []Feature{
	{
		Name:        "动画拓展",
		Description: "增加角色动作、攻击动画和过渡效果，让游戏表现更加生动。",
		Hashes: []string{
			"cbf33a7187aaa36f1920bc21c6bcfe65f8e1cbd7338b7357e960d63552804d2b",
			"07ff1b184ad9fcb9c540d3bb8d004d903474f90c195e4e3ce00be56faac243b8",
		},
	},
	{
		Name:        "氛围优化",
		Description: "改善渲染效果与画面细节，提升整体氛围。",
		Hashes: []string{
			"bf30dcef0104da257fef756463d90820929efcaae6c0f1d34589b383fc420340",
			"14d44503562adc0228b0dcbbe2532b1af6d55afe9930e9565c574b3347bc70b7",
			"60e0f15e1174bd618d1dee2f2ce6bedba0a16f154757dbad145eba965e46e824",
			"b42bd00c7ab6e044d21f845dedc58a4524e45643f3685bf1c2343098e74d662e",
			"6497282f734e876207fc5824421bafe404b32aab1611435fcd046b9ba1d2fdb8",
			"7e9cf735b26d2d5ea8d1cc247dd60351be8bd401b30556b84054df85d334f9c4",
			"33e8acb7f264f691f2e6cf7c0d7b9152b54cf0353f37fc8a00d7312f26b36e3e",
		},
	},
	{
		Name:        "性能优化",
		Description: "改善游戏性能，提升运行效率和稳定性，推荐保持开启。",
		Hashes: []string{
			"c161d70a670dcade5d04a38e33cf0aefb73ce70f74b2576a4d40d480a3389e1e",
			"b3d76f1deb9adc4012721ed56f9869d49da7b0c0e57b42a85029fe291ac200ca",
			"52eb05524215427a8bdf53212b8fc2361b6a365607a36e94773fea55ddaf862a",
			"9b4443f072ff54288d51466b8dc3f327490848ebb5b75ee8755890949aa1f4a7",
			"aea0eefb174d19ad97a812b6de2e304f0d3abc50357b316908cf1035f1514b07",
		},
	},
}

const modsDir = "mods"

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

func sha256File(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}

func isModFile(name string) bool {
	return strings.HasSuffix(name, ".jar") || strings.HasSuffix(name, ".jar.disabled")
}

// buildHashIndex maps sha256 -> absolute file path for every jar/jar.disabled in mods/.
func buildHashIndex() map[string]string {
	index := map[string]string{}

	entries, err := os.ReadDir(modsDir)
	if err != nil {
		return index
	}

	for _, entry := range entries {
		if entry.IsDir() || !isModFile(entry.Name()) {
			continue
		}

		path := filepath.Join(modsDir, entry.Name())

		hash, err := sha256File(path)
		if err != nil {
			continue
		}

		index[hash] = path
	}

	return index
}

// displayName extracts a readable label from a mod filename, preferring a
// trailing Chinese tag if the filename carries one (e.g. "sodium-...jar-钠").
func displayName(path string) string {
	name := filepath.Base(path)
	name = strings.TrimSuffix(name, ".disabled")
	name = strings.TrimSuffix(name, ".jar")

	if idx := strings.LastIndex(name, "-"); idx != -1 {
		tail := name[idx+1:]
		if containsHan(tail) {
			return tail
		}
	}

	return name
}

func containsHan(s string) bool {
	for _, r := range s {
		if unicode.Is(unicode.Han, r) {
			return true
		}
	}
	return false
}

// GetFeatureStatuses returns the render model for every feature card.
func (a *App) GetFeatureStatuses() []FeatureStatus {
	index := buildHashIndex()
	result := make([]FeatureStatus, 0, len(features))

	for _, feature := range features {
		status := FeatureStatus{
			Name:        feature.Name,
			Description: feature.Description,
			Mods:        make([]ModStatus, 0, len(feature.Hashes)),
		}

		for _, hash := range feature.Hashes {
			path, found := index[hash]

			if !found {
				status.Mods = append(status.Mods, ModStatus{
					DisplayName: "（未找到对应文件）",
					Enabled:     false,
					Found:       false,
				})
				continue
			}

			enabled := strings.HasSuffix(path, ".jar")

			if enabled {
				status.Enabled = true
			}

			status.Mods = append(status.Mods, ModStatus{
				DisplayName: displayName(path),
				Enabled:     enabled,
				Found:       true,
			})
		}

		result = append(result, status)
	}

	return result
}

// ToggleFeature enables or disables every mod belonging to the feature at
// featureIndex by renaming between .jar and .jar.disabled.
func (a *App) ToggleFeature(featureIndex int, enabled bool) []FeatureStatus {
	if featureIndex < 0 || featureIndex >= len(features) {
		return a.GetFeatureStatuses()
	}

	index := buildHashIndex()

	for _, hash := range features[featureIndex].Hashes {
		path, found := index[hash]
		if !found {
			continue
		}

		if enabled && strings.HasSuffix(path, ".jar.disabled") {
			os.Rename(path, strings.TrimSuffix(path, ".disabled"))
		} else if !enabled && strings.HasSuffix(path, ".jar") {
			os.Rename(path, path+".disabled")
		}
	}

	return a.GetFeatureStatuses()
}
