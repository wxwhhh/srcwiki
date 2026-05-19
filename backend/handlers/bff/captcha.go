package handlers

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"math"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// CaptchaStore 验证码存储（内存，带过期）
type captchaEntry struct {
	code     string
	expireAt time.Time
}

var (
	captchaStore   = make(map[string]captchaEntry)
	captchaStoreMu sync.Mutex
	captchaChars   = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789" // 大写字母 + 数字，去掉容易混淆的字符(0/O/1/I/l)
)

func init() {
	// 定期清理过期验证码
	go func() {
		for {
			time.Sleep(1 * time.Minute)
			captchaStoreMu.Lock()
			now := time.Now()
			for k, v := range captchaStore {
				if now.After(v.expireAt) {
					delete(captchaStore, k)
				}
			}
			captchaStoreMu.Unlock()
		}
	}()
}

// generateCaptchaCode 生成 4 位验证码
func generateCaptchaCode() string {
	b := make([]byte, 4)
	for i := range b {
		b[i] = captchaChars[rand.Intn(len(captchaChars))]
	}
	return string(b)
}

// GenerateCaptcha 生成验证码图片
func GenerateCaptcha(c *gin.Context) {
	code := generateCaptchaCode()
	captchaID := generateCaptchaCode() + generateCaptchaCode() // 8位随机ID

	// 存储验证码，5 分钟过期
	captchaStoreMu.Lock()
	captchaStore[captchaID] = captchaEntry{
		code:     code,
		expireAt: time.Now().Add(5 * time.Minute),
	}
	captchaStoreMu.Unlock()

	// 生成图片
	img := renderCaptcha(code)
	var buf bytes.Buffer
	png.Encode(&buf, img)

	c.JSON(200, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"captcha_id":  captchaID,
			"captcha_img": "data:image/png;base64," + base64Encode(buf.Bytes()),
		},
	})
}

// VerifyCaptcha 验证验证码
func VerifyCaptcha(captchaID, captchaCode string) bool {
	if captchaID == "" || captchaCode == "" {
		return false
	}
	captchaStoreMu.Lock()
	defer captchaStoreMu.Unlock()

	entry, ok := captchaStore[captchaID]
	if !ok {
		return false
	}
	delete(captchaStore, captchaID) // 一次性，用完即删

	if time.Now().After(entry.expireAt) {
		return false
	}
	return strings.EqualFold(entry.code, captchaCode)
}

// base64Encode 简单的 base64 编码
func base64Encode(data []byte) string {
	const base64Table = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"
	buf := make([]byte, 0, (len(data)+2)/3*4)
	for i := 0; i < len(data); i += 3 {
		var b0, b1, b2 byte
		b0 = data[i]
		if i+1 < len(data) {
			b1 = data[i+1]
		}
		if i+2 < len(data) {
			b2 = data[i+2]
		}
		buf = append(buf, base64Table[b0>>2])
		buf = append(buf, base64Table[((b0&0x03)<<4)|(b1>>4)])
		if i+1 < len(data) {
			buf = append(buf, base64Table[((b1&0x0f)<<2)|(b2>>6)])
		} else {
			buf = append(buf, '=')
		}
		if i+2 < len(data) {
			buf = append(buf, base64Table[b2&0x3f])
		} else {
			buf = append(buf, '=')
		}
	}
	return string(buf)
}

// renderCaptcha 渲染验证码图片（120x40，彩色背景 + 噪点 + 扭曲字符）
func renderCaptcha(code string) *image.RGBA {
	w, h := 120, 40
	img := image.NewRGBA(image.Rect(0, 0, w, h))

	// 彩色随机背景
	bgR := uint8(rand.Intn(50) + 200)
	bgG := uint8(rand.Intn(50) + 200)
	bgB := uint8(rand.Intn(50) + 200)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, color.RGBA{bgR, bgG, bgB, 255})
		}
	}

	// 噪点线（3 条，颜色较淡）
	for i := 0; i < 3; i++ {
		x1, y1 := rand.Intn(w), rand.Intn(h)
		x2, y2 := rand.Intn(w), rand.Intn(h)
		drawLine(img, x1, y1, x2, y2, color.RGBA{
			uint8(rand.Intn(100) + 150),
			uint8(rand.Intn(100) + 150),
			uint8(rand.Intn(100) + 150),
			180,
		})
	}

	// 噪点（30 个，较淡）
	for i := 0; i < 30; i++ {
		x, y := rand.Intn(w), rand.Intn(h)
		img.Set(x, y, color.RGBA{
			uint8(rand.Intn(100) + 150),
			uint8(rand.Intn(100) + 150),
			uint8(rand.Intn(100) + 150),
			uint8(rand.Intn(64) + 64),
		})
	}

	// 绘制扭曲字符
	charWidth := w / 5
	for i, ch := range code {
		x := i*charWidth + 4
		// 随机偏移
		y := rand.Intn(6) + 14
		// 随机颜色（深色）
		clr := color.RGBA{
			uint8(rand.Intn(100)),
			uint8(rand.Intn(100)),
			uint8(rand.Intn(100)),
			255,
		}
		// 随机旋转角度（-20° ~ +20°）
		angle := (rand.Float64() - 0.5) * 0.7
		drawChar(img, ch, x, y, 22+rand.Intn(4), angle, clr)
	}

	return img
}

// drawLine 画直线（Bresenham）
func drawLine(img *image.RGBA, x0, y0, x1, y1 int, clr color.RGBA) {
	dx := abs(x1 - x0)
	dy := -abs(y1 - y0)
	sx := 1
	if x0 > x1 {
		sx = -1
	}
	sy := 1
	if y0 > y1 {
		sy = -1
	}
	err := dx + dy
	for {
		if x0 >= 0 && x0 < img.Bounds().Dx() && y0 >= 0 && y0 < img.Bounds().Dy() {
			img.Set(x0, y0, clr)
		}
		if x0 == x1 && y0 == y1 {
			break
		}
		e2 := 2 * err
		if e2 >= dy {
			err += dy
			x0 += sx
		}
		if e2 <= dx {
			err += dx
			y0 += sy
		}
	}
}

// drawChar 绘制字符（简易 5x7 点阵 + 旋转）
func drawChar(img *image.RGBA, ch rune, x, y, size int, angle float64, clr color.RGBA) {
	// 5x7 点阵字体数据
	patterns := map[rune][7]byte{
		'A': {0x04, 0x0A, 0x11, 0x1F, 0x11, 0x11, 0x11},
		'B': {0x1E, 0x11, 0x11, 0x1E, 0x11, 0x11, 0x1E},
		'C': {0x0E, 0x11, 0x10, 0x10, 0x10, 0x11, 0x0E},
		'D': {0x1C, 0x12, 0x11, 0x11, 0x11, 0x12, 0x1C},
		'E': {0x1F, 0x10, 0x10, 0x1E, 0x10, 0x10, 0x1F},
		'F': {0x1F, 0x10, 0x10, 0x1E, 0x10, 0x10, 0x10},
		'G': {0x0E, 0x11, 0x10, 0x17, 0x11, 0x11, 0x0E},
		'H': {0x11, 0x11, 0x11, 0x1F, 0x11, 0x11, 0x11},
		'J': {0x07, 0x02, 0x02, 0x02, 0x02, 0x12, 0x0C},
		'K': {0x11, 0x12, 0x14, 0x18, 0x14, 0x12, 0x11},
		'L': {0x10, 0x10, 0x10, 0x10, 0x10, 0x10, 0x1F},
		'M': {0x11, 0x1B, 0x15, 0x15, 0x11, 0x11, 0x11},
		'N': {0x11, 0x19, 0x15, 0x13, 0x11, 0x11, 0x11},
		'P': {0x1E, 0x11, 0x11, 0x1E, 0x10, 0x10, 0x10},
		'Q': {0x0E, 0x11, 0x11, 0x11, 0x15, 0x12, 0x0D},
		'R': {0x1E, 0x11, 0x11, 0x1E, 0x14, 0x12, 0x11},
		'S': {0x0E, 0x11, 0x10, 0x0E, 0x01, 0x11, 0x0E},
		'T': {0x1F, 0x04, 0x04, 0x04, 0x04, 0x04, 0x04},
		'U': {0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x0E},
		'V': {0x11, 0x11, 0x11, 0x11, 0x0A, 0x0A, 0x04},
		'W': {0x11, 0x11, 0x11, 0x15, 0x15, 0x1B, 0x11},
		'X': {0x11, 0x11, 0x0A, 0x04, 0x0A, 0x11, 0x11},
		'Y': {0x11, 0x11, 0x0A, 0x04, 0x04, 0x04, 0x04},
		'Z': {0x1F, 0x01, 0x02, 0x04, 0x08, 0x10, 0x1F},
		'a': {0x00, 0x00, 0x0E, 0x01, 0x0F, 0x11, 0x0F},
		'b': {0x10, 0x10, 0x16, 0x19, 0x11, 0x11, 0x1E},
		'c': {0x00, 0x00, 0x0E, 0x10, 0x10, 0x11, 0x0E},
		'd': {0x01, 0x01, 0x0D, 0x13, 0x11, 0x11, 0x0F},
		'e': {0x00, 0x00, 0x0E, 0x11, 0x1F, 0x10, 0x0E},
		'f': {0x06, 0x09, 0x08, 0x1C, 0x08, 0x08, 0x08},
		'g': {0x00, 0x0F, 0x11, 0x11, 0x0F, 0x01, 0x0E},
		'h': {0x10, 0x10, 0x16, 0x19, 0x11, 0x11, 0x11},
		'j': {0x02, 0x00, 0x06, 0x02, 0x02, 0x12, 0x0C},
		'k': {0x10, 0x10, 0x12, 0x14, 0x18, 0x14, 0x12},
		'm': {0x00, 0x00, 0x1A, 0x15, 0x15, 0x11, 0x11},
		'n': {0x00, 0x00, 0x16, 0x19, 0x11, 0x11, 0x11},
		'p': {0x00, 0x00, 0x16, 0x19, 0x1E, 0x10, 0x10},
		'q': {0x00, 0x00, 0x0D, 0x13, 0x0F, 0x01, 0x01},
		'r': {0x00, 0x00, 0x16, 0x19, 0x10, 0x10, 0x10},
		's': {0x00, 0x00, 0x0E, 0x10, 0x0E, 0x01, 0x1E},
		't': {0x08, 0x08, 0x1C, 0x08, 0x08, 0x09, 0x06},
		'u': {0x00, 0x00, 0x11, 0x11, 0x11, 0x13, 0x0D},
		'v': {0x00, 0x00, 0x11, 0x11, 0x11, 0x0A, 0x04},
		'w': {0x00, 0x00, 0x11, 0x11, 0x15, 0x15, 0x0A},
		'x': {0x00, 0x00, 0x11, 0x0A, 0x04, 0x0A, 0x11},
		'y': {0x00, 0x00, 0x11, 0x11, 0x0F, 0x01, 0x0E},
		'z': {0x00, 0x00, 0x1F, 0x02, 0x04, 0x08, 0x1F},
		'0': {0x0E, 0x11, 0x13, 0x15, 0x19, 0x11, 0x0E},
		'1': {0x04, 0x0C, 0x04, 0x04, 0x04, 0x04, 0x0E},
		'2': {0x0E, 0x11, 0x01, 0x06, 0x08, 0x10, 0x1F},
		'3': {0x0E, 0x11, 0x01, 0x06, 0x01, 0x11, 0x0E},
		'4': {0x02, 0x06, 0x0A, 0x12, 0x1F, 0x02, 0x02},
		'5': {0x1F, 0x10, 0x1E, 0x01, 0x01, 0x11, 0x0E},
		'6': {0x06, 0x08, 0x10, 0x1E, 0x11, 0x11, 0x0E},
		'7': {0x1F, 0x01, 0x02, 0x04, 0x08, 0x08, 0x08},
		'8': {0x0E, 0x11, 0x11, 0x0E, 0x11, 0x11, 0x0E},
		'9': {0x0E, 0x11, 0x11, 0x0F, 0x01, 0x02, 0x0C},
	}

	pattern, ok := patterns[ch]
	if !ok {
		pattern = patterns['0'] // fallback
	}

	scale := float64(size) / 7.0
	cx := float64(x) + float64(size)*0.5
	cy := float64(y) + float64(size)*0.5

	for row := 0; row < 7; row++ {
		for col := 0; col < 5; col++ {
			if pattern[row]&(1<<(4-col)) != 0 {
				// 计算旋转后的坐标
				fx := float64(col) - 2.0
				fy := float64(row) - 3.0
				rx := fx*math.Cos(angle) - fy*math.Sin(angle)
				ry := fx*math.Sin(angle) + fy*math.Cos(angle)
				px := int(cx + rx*scale)
				py := int(cy + ry*scale)

				// 画像素块
				for dy := 0; dy < int(scale); dy++ {
					for dx := 0; dx < int(scale); dx++ {
						xx, yy := px+dx, py+dy
						if xx >= 0 && xx < img.Bounds().Dx() && yy >= 0 && yy < img.Bounds().Dy() {
							img.Set(xx, yy, clr)
						}
					}
				}
			}
		}
	}
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
