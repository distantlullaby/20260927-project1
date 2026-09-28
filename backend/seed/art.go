package seed

import (
	"fmt"
	"os"
	"path/filepath"
)

// Scene 演示插画的场景类型
type Scene int

const (
	SceneAlley Scene = iota
	SceneSchool
	SceneRiverside
)

func fileURL(uploadDir, filename string) string {
	return "/uploads/" + filename
}

// writeSVG 写入一张 SVG 并返回可访问 URL
func writeSVG(uploadDir, filename, svg string) string {
	_ = os.MkdirAll(uploadDir, 0o755)
	path := filepath.Join(uploadDir, filename)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		_ = os.WriteFile(path, []byte(svg), 0o644)
	}
	return fileURL(uploadDir, filename)
}

// OldPhotoURL 生成泛黄的“过去”老照片
func OldPhotoURL(dir, name string, scene Scene) string {
	return writeSVG(dir, name+".svg", vintageSVG(scene, true))
}

// NewPhotoURL 生成明亮的“当下”新照片
func NewPhotoURL(dir, name string, scene Scene) string {
	return writeSVG(dir, name+".svg", vintageSVG(scene, false))
}

func vintageSVG(scene Scene, old bool) string {
	var bg, top, bottom, label, art string
	if old {
		bg, top, bottom = "#f0e2c0", "#e7d3a8", "#d9bf90"
		label = "过去 · 老照片"
	} else {
		bg, top, bottom = "#eaf6f0", "#cfe8dd", "#a9d6c4"
		label = "当下 · 现场"
	}

	switch scene {
	case SceneSchool:
		art = `
<rect x="70" y="150" width="260" height="90" fill="#c98d5e"/>
<rect x="95" y="175" width="34" height="34" fill="#7fd4d9"/>
<rect x="155" y="175" width="34" height="34" fill="#7fd4d9"/>
<rect x="215" y="175" width="34" height="34" fill="#7fd4d9"/>
<rect x="275" y="175" width="30" height="34" fill="#8f6a48"/>
<polygon points="60,150 200,95 340,150" fill="#a9673d"/>
<rect x="120" y="240" width="160" height="40" rx="4" fill="#6fae6a"/>
<circle cx="300" cy="95" r="22" fill="#f4c95d"/>`
		if old {
			label = "过去 · 三中门口"
		} else {
			label = "当下 · 三中门口"
		}
	case SceneRiverside:
		art = `
<rect x="0" y="230" width="400" height="70" fill="#8fb8d6"/>
<rect x="0" y="200" width="400" height="34" fill="#b9a06a"/>
<rect x="30" y="180" width="340" height="22" fill="#cbb37c"/>
<line x1="70" y1="180" x2="120" y2="110" stroke="#8c7a52" stroke-width="4"/>
<line x1="120" y1="110" x2="150" y2="135" stroke="#555" stroke-width="2"/>
<polygon points="150,135 182,128 152,150" fill="#d96b54"/>
<line x1="300" y1="180" x2="268" y2="120" stroke="#8c7a52" stroke-width="4"/>
<line x1="268" y1="120" x2="300" y2="130" stroke="#555" stroke-width="2"/>
<polygon points="300,130 330,120 302,146" fill="#4f7fb0"/>
<circle cx="330" cy="80" r="20" fill="#f4c95d"/>`
		if old {
			label = "过去 · 长江大堤"
		} else {
			label = "当下 · 长江大堤"
		}
	default: // SceneAlley 老槐树
		art = `
<rect x="0" y="250" width="400" height="50" fill="#c8b58d"/>
<rect x="40" y="120" width="70" height="130" fill="#dcc39b"/>
<rect x="290" y="100" width="70" height="150" fill="#d3b98f"/>
<rect x="182" y="170" width="36" height="80" fill="#8a5a3b"/>
<circle cx="200" cy="140" r="70" fill="#7fae5f"/>
<circle cx="155" cy="115" r="38" fill="#93bd72"/>
<circle cx="246" cy="118" r="36" fill="#8bb768"/>
<circle cx="305" cy="70" r="20" fill="#f4c95d"/>`
		if old {
			label = "过去 · 槐树巷"
		} else {
			label = "当下 · 槐树巷"
		}
	}

	stamp := ""
	if old {
		stamp = `<rect x="320" y="20" width="52" height="40" fill="none" stroke="#b08a55" stroke-width="2" stroke-dasharray="4 3"/>
<circle cx="346" cy="40" r="12" fill="none" stroke="#b08a55" stroke-width="2"/>`
	}

	return fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" width="400" height="300" viewBox="0 0 400 300">
<rect width="400" height="300" fill="%s"/>
<rect x="12" y="12" width="376" height="276" fill="none" stroke="%s" stroke-width="2"/>
%s
%s
<rect x="0" y="250" width="400" height="50" fill="%s" opacity="0.92"/>
<text x="24" y="282" font-family="Georgia,'Songti SC',serif" font-size="18" fill="#5a4632">%s</text>
</svg>`, bg, top, art, stamp, bottom, label)
}
