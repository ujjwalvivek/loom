package renderer

import (
	"github.com/go-gl/gl/v3.3-core/gl"
)

type Batcher struct {
	vao           uint32
	vbo           uint32
	ebo           uint32
	vertices      []float32
	maxQuads      int
	quadsCount    int
	activeTexture uint32
	activeShader  *Shader
	solidTex      *Texture
}

func NewBatcher(maxQuads int, solidTex *Texture) *Batcher {
	b := &Batcher{
		maxQuads:      maxQuads,
		vertices:      make([]float32, maxQuads*4*8), // 4 vertices, 8 floats per vertex [x, y, u, v, r, g, b, a]
		solidTex:      solidTex,
		activeTexture: solidTex.ID,
	}

	gl.GenVertexArrays(1, &b.vao)
	gl.GenBuffers(1, &b.vbo)
	gl.GenBuffers(1, &b.ebo)

	gl.BindVertexArray(b.vao)

	gl.BindBuffer(gl.ARRAY_BUFFER, b.vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(b.vertices)*4, gl.Ptr(b.vertices), gl.DYNAMIC_DRAW)

	indices := make([]uint32, maxQuads*6)
	for i := 0; i < maxQuads; i++ {
		offset := uint32(i * 4)
		indices[i*6+0] = offset + 0
		indices[i*6+1] = offset + 1
		indices[i*6+2] = offset + 2
		indices[i*6+3] = offset + 2
		indices[i*6+4] = offset + 3
		indices[i*6+5] = offset + 0
	}
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, b.ebo)
	gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, len(indices)*4, gl.Ptr(indices), gl.STATIC_DRAW)

	// Layout Stride: 8 floats * 4 bytes = 32 bytes
	stride := int32(8 * 4)

	// Position (Location 0, Vec2)
	gl.VertexAttribPointer(0, 2, gl.FLOAT, false, stride, gl.PtrOffset(0))
	gl.EnableVertexAttribArray(0)

	// UV (Location 1, Vec2)
	gl.VertexAttribPointer(1, 2, gl.FLOAT, false, stride, gl.PtrOffset(2*4))
	gl.EnableVertexAttribArray(1)

	// Color (Location 2, Vec4)
	gl.VertexAttribPointer(2, 4, gl.FLOAT, false, stride, gl.PtrOffset(4*4))
	gl.EnableVertexAttribArray(2)

	gl.BindVertexArray(0)

	return b
}

func (b *Batcher) DrawQuad(x, y, w, h float32, u0, v0, u1, v1 float32, r, g, bVal, a float32, textureID uint32, shader *Shader) {
	if b.quadsCount >= b.maxQuads || textureID != b.activeTexture || b.activeShader == nil || shader.ID != b.activeShader.ID {
		b.Flush()
	}

	b.activeTexture = textureID
	b.activeShader = shader

	idx := b.quadsCount * 32

	// V0: Top-Left
	b.vertices[idx+0] = x
	b.vertices[idx+1] = y
	b.vertices[idx+2] = u0
	b.vertices[idx+3] = v0
	b.vertices[idx+4] = r
	b.vertices[idx+5] = g
	b.vertices[idx+6] = bVal
	b.vertices[idx+7] = a

	// V1: Top-Right
	b.vertices[idx+8] = x + w
	b.vertices[idx+9] = y
	b.vertices[idx+10] = u1
	b.vertices[idx+11] = v0
	b.vertices[idx+12] = r
	b.vertices[idx+13] = g
	b.vertices[idx+14] = bVal
	b.vertices[idx+15] = a

	// V2: Bottom-Right
	b.vertices[idx+16] = x + w
	b.vertices[idx+17] = y + h
	b.vertices[idx+18] = u1
	b.vertices[idx+19] = v1
	b.vertices[idx+20] = r
	b.vertices[idx+21] = g
	b.vertices[idx+22] = bVal
	b.vertices[idx+23] = a

	// V3: Bottom-Left
	b.vertices[idx+24] = x
	b.vertices[idx+25] = y + h
	b.vertices[idx+26] = u0
	b.vertices[idx+27] = v1
	b.vertices[idx+28] = r
	b.vertices[idx+29] = g
	b.vertices[idx+30] = bVal
	b.vertices[idx+31] = a

	b.quadsCount++
}

func (b *Batcher) Flush() {
	if b.quadsCount == 0 {
		return
	}

	b.activeShader.Use()
	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, b.activeTexture)
	b.activeShader.SetUniform1i("u_Texture", 0)

	gl.BindBuffer(gl.ARRAY_BUFFER, b.vbo)
	// BufferSubData ensures we only send the written chunk of vertices to GPU
	gl.BufferSubData(gl.ARRAY_BUFFER, 0, b.quadsCount*32*4, gl.Ptr(b.vertices))

	gl.BindVertexArray(b.vao)
	gl.DrawElements(gl.TRIANGLES, int32(b.quadsCount*6), gl.UNSIGNED_INT, gl.PtrOffset(0))
	gl.BindVertexArray(0)

	b.quadsCount = 0
}

func (b *Batcher) Delete() {
	gl.DeleteVertexArrays(1, &b.vao)
	gl.DeleteBuffers(1, &b.vbo)
	gl.DeleteBuffers(1, &b.ebo)
}
