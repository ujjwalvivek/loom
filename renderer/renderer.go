package renderer

import (
	"log"
	"math"

	"github.com/go-gl/gl/v3.3-core/gl"
	loommath "github.com/ujjwalvivek/loom/math"
)

type SpriteOptions struct {
	U0, V0, U1, V1 float32
	FlipX          bool
	FlipY          bool
	Angle          float32
}

type RenderSystem struct {
	width       int32
	height      int32
	batcher     *Batcher
	postfx      *PostFX
	Camera      *Camera
	solidTex    *Texture
	radialTex   *Texture
	noiseTex    *Texture
	batchShader *Shader
}

func NewRenderSystem(width, height int32) *RenderSystem {
	// Create a 1x1 solid white texture for drawing solid colored shapes
	solidTex := GenerateSolidTexture(loommath.Color{R: 1, G: 1, B: 1, A: 1})
	radialTex := GenerateRadialGradient(256)
	noiseTex := GenerateNoiseTexture(256, 12345)

	// Compile batch rendering shaders
	shader, err := NewShader(batchVertexShader, batchFragmentShader)
	if err != nil {
		log.Fatalln("Failed to compile render system batch shader:", err)
	}

	batcher := NewBatcher(8192, solidTex)
	postfx := NewPostFX(width, height)
	camera := NewCamera()
	camera.Pos = loommath.Vec2{X: float32(width) / 2.0, Y: float32(height) / 2.0}

	return &RenderSystem{
		width:       width,
		height:      height,
		batcher:     batcher,
		postfx:      postfx,
		Camera:      camera,
		solidTex:    solidTex,
		radialTex:   radialTex,
		noiseTex:    noiseTex,
		batchShader: shader,
	}
}

func (rs *RenderSystem) Begin() {
	rs.postfx.Begin()
	gl.ClearColor(0.0, 0.0, 0.0, 1.0)
	gl.Clear(gl.COLOR_BUFFER_BIT)
	
	// Upload camera projection matrix to batch shader
	rs.batchShader.Use()
	proj := rs.Camera.GetProjectionMatrix(float32(rs.width), float32(rs.height))
	rs.batchShader.SetUniformMat4("u_Proj", &proj)
}

func (rs *RenderSystem) End(time float32) {
	rs.batcher.Flush()
	rs.postfx.End(time)
}

func (rs *RenderSystem) Update(dt float32) {
	rs.Camera.Update(dt)
}

func (rs *RenderSystem) Clear(color loommath.Color) {
	gl.ClearColor(color.R, color.G, color.B, color.A)
	gl.Clear(gl.COLOR_BUFFER_BIT)
}

func (rs *RenderSystem) Rect(pos, size loommath.Vec2, color loommath.Color) {
	rs.batcher.DrawQuad(
		pos.X, pos.Y, size.X, size.Y,
		0, 0, 1, 1,
		color.R, color.G, color.B, color.A,
		rs.solidTex.ID,
		rs.batchShader,
	)
}

func (rs *RenderSystem) Sprite(pos, size loommath.Vec2, textureID uint32, options SpriteOptions) {
	u0, v0, u1, v1 := options.U0, options.V0, options.U1, options.V1
	if u1 == 0 && v1 == 0 {
		u1, v1 = 1, 1
	}

	if options.FlipX {
		u0, u1 = u1, u0
	}
	if options.FlipY {
		v0, v1 = v1, v0
	}

	// Simple check: if textureID is 0, default to solid texture
	tid := textureID
	if tid == 0 {
		tid = rs.solidTex.ID
	}

	rs.batcher.DrawQuad(
		pos.X, pos.Y, size.X, size.Y,
		u0, v0, u1, v1,
		1, 1, 1, 1,
		tid,
		rs.batchShader,
	)
}

func (rs *RenderSystem) Circle(pos loommath.Vec2, radius float32, color loommath.Color) {
	// Draw using pre-generated radial texture
	rs.batcher.DrawQuad(
		pos.X-radius, pos.Y-radius, radius*2, radius*2,
		0, 0, 1, 1,
		color.R, color.G, color.B, color.A,
		rs.radialTex.ID,
		rs.batchShader,
	)
}

func (rs *RenderSystem) Line(a, b loommath.Vec2, color loommath.Color, thickness float32) {
	dx := b.X - a.X
	dy := b.Y - a.Y
	length := float32(math.Sqrt(float64(dx*dx + dy*dy)))
	if length == 0 {
		return
	}

	rs.batcher.DrawQuad(
		a.X, a.Y-thickness/2.0, length, thickness,
		0, 0, 1, 1,
		color.R, color.G, color.B, color.A,
		rs.solidTex.ID,
		rs.batchShader,
	)
}

func (rs *RenderSystem) GetSolidTexture() *Texture {
	return rs.solidTex
}

func (rs *RenderSystem) GetRadialTexture() *Texture {
	return rs.radialTex
}

func (rs *RenderSystem) GetNoiseTexture() *Texture {
	return rs.noiseTex
}

func (rs *RenderSystem) SetUISpace(uiSpace bool) {
	// Flush current batch to push world quads before we change projection
	rs.batcher.Flush()
	
	rs.batchShader.Use()
	var proj [16]float32
	if uiSpace {
		// Centered camera representing static screen coordinates (0 to width, 0 to height)
		tempCam := &Camera{
			Pos:  loommath.Vec2{X: float32(rs.width) / 2.0, Y: float32(rs.height) / 2.0},
			Zoom: 1.0,
		}
		proj = tempCam.GetProjectionMatrix(float32(rs.width), float32(rs.height))
	} else {
		proj = rs.Camera.GetProjectionMatrix(float32(rs.width), float32(rs.height))
	}
	rs.batchShader.SetUniformMat4("u_Proj", &proj)
}

func (rs *RenderSystem) Delete() {
	rs.batcher.Delete()
	rs.postfx.Delete()
	rs.solidTex.Delete()
	rs.radialTex.Delete()
	rs.noiseTex.Delete()
	rs.batchShader.Delete()
}

// Shader definitions
const batchVertexShader = `
#version 330 core
layout (location = 0) in vec2 aPos;
layout (location = 1) in vec2 aTexCoords;
layout (location = 2) in vec4 aColor;

out vec2 TexCoords;
out vec4 Color;

uniform mat4 u_Proj;

void main() {
    TexCoords = aTexCoords;
    Color = aColor;
    gl_Position = u_Proj * vec4(aPos, 0.0, 1.0);
}
` + "\x00"

const batchFragmentShader = `
#version 330 core
out vec4 FragColor;

in vec2 TexCoords;
in vec4 Color;

uniform sampler2D u_Texture;

void main() {
    FragColor = texture(u_Texture, TexCoords) * Color;
}
` + "\x00"
