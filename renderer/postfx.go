package renderer

import (
	"log"

	"github.com/go-gl/gl/v3.3-core/gl"
)

type PostFX struct {
	width  int32
	height int32

	sceneFBO uint32
	sceneTex uint32

	pingFBO [2]uint32
	pingTex [2]uint32

	extractShader   *Shader
	blurShader      *Shader
	compositeShader *Shader

	quadVAO uint32
	quadVBO uint32
}

func NewPostFX(width, height int32) *PostFX {
	p := &PostFX{
		width:  width,
		height: height,
	}

	p.setupFramebuffers()
	p.setupShaders()
	p.quadVAO, p.quadVBO = createScreenQuad()

	return p
}

func (p *PostFX) setupFramebuffers() {
	// Scene FBO (draw target for entire game frame)
	gl.GenFramebuffers(1, &p.sceneFBO)
	gl.BindFramebuffer(gl.FRAMEBUFFER, p.sceneFBO)

	gl.GenTextures(1, &p.sceneTex)
	gl.BindTexture(gl.TEXTURE_2D, p.sceneTex)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, p.width, p.height, 0, gl.RGBA, gl.UNSIGNED_BYTE, nil)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, p.sceneTex, 0)

	if gl.CheckFramebufferStatus(gl.FRAMEBUFFER) != gl.FRAMEBUFFER_COMPLETE {
		log.Fatalln("Error: Scene Framebuffer is incomplete")
	}

	// Ping-Pong FBOs (blurred buffers at half size to optimize blur execution speed)
	gl.GenFramebuffers(2, &p.pingFBO[0])
	gl.GenTextures(2, &p.pingTex[0])

	for i := 0; i < 2; i++ {
		gl.BindFramebuffer(gl.FRAMEBUFFER, p.pingFBO[i])
		gl.BindTexture(gl.TEXTURE_2D, p.pingTex[i])
		gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, p.width/2, p.height/2, 0, gl.RGBA, gl.UNSIGNED_BYTE, nil)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
		gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, p.pingTex[i], 0)

		if gl.CheckFramebufferStatus(gl.FRAMEBUFFER) != gl.FRAMEBUFFER_COMPLETE {
			log.Fatalln("Error: PingPong Framebuffer", i, "is incomplete")
		}
	}

	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
}

func (p *PostFX) setupShaders() {
	var err error
	
	// Brightness extraction shader
	p.extractShader, err = NewShader(vertexShaderSource, extractShaderSource)
	if err != nil {
		log.Fatalln("Failed compiling PostFX Extract Shader:", err)
	}

	// Blur shader
	p.blurShader, err = NewShader(vertexShaderSource, blurShaderSource)
	if err != nil {
		log.Fatalln("Failed compiling PostFX Blur Shader:", err)
	}

	// Composite pass shader
	p.compositeShader, err = NewShader(vertexShaderSource, compositeShaderSource)
	if err != nil {
		log.Fatalln("Failed compiling PostFX Composite Shader:", err)
	}
}

func (p *PostFX) Begin() {
	gl.BindFramebuffer(gl.FRAMEBUFFER, p.sceneFBO)
	gl.Viewport(0, 0, p.width, p.height)
}

func (p *PostFX) End(time float32) {
	// Unbind framebuffer to return rendering directly to screen buffer
	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
	gl.Viewport(0, 0, p.width, p.height)

	// Brightness Extraction
	gl.BindFramebuffer(gl.FRAMEBUFFER, p.pingFBO[0])
	gl.Viewport(0, 0, p.width/2, p.height/2)
	gl.Clear(gl.COLOR_BUFFER_BIT)
	
	p.extractShader.Use()
	p.extractShader.SetUniform1f("u_Threshold", 0.85)
	
	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, p.sceneTex)
	p.extractShader.SetUniform1i("u_Scene", 0)

	gl.BindVertexArray(p.quadVAO)
	gl.DrawArrays(gl.TRIANGLES, 0, 6)

	// Ping-pong Gaussian Blur (horizontal/vertical swapping passes)
	horizontal := true
	p.blurShader.Use()
	
	for i := 0; i < 4; i++ {
		gl.BindFramebuffer(gl.FRAMEBUFFER, p.pingFBO[toInt(horizontal)])
		p.blurShader.SetUniform1i("u_Horizontal", toInt32(horizontal))
		
		gl.ActiveTexture(gl.TEXTURE0)
		gl.BindTexture(gl.TEXTURE_2D, p.pingTex[toInt(!horizontal)])
		p.blurShader.SetUniform1i("u_Image", 0)
		
		gl.DrawArrays(gl.TRIANGLES, 0, 6)
		horizontal = !horizontal
	}

	// Final Composite pass
	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
	gl.Viewport(0, 0, p.width, p.height)
	gl.Clear(gl.COLOR_BUFFER_BIT)

	p.compositeShader.Use()
	p.compositeShader.SetUniform1f("u_Time", time)
	
	// Bloom
	p.compositeShader.SetUniform1i("u_BloomEnabled", 1)
	p.compositeShader.SetUniform1f("u_BloomIntensity", 0.71)

	// Vignette
	p.compositeShader.SetUniform1i("u_VignetteEnabled", 0)
	p.compositeShader.SetUniform1f("u_VignetteIntensity", 0.0)
	p.compositeShader.SetUniform1f("u_VignetteRoundness", 0.0)

	// Film Grain
	p.compositeShader.SetUniform1i("u_GrainEnabled", 0)
	p.compositeShader.SetUniform1f("u_GrainIntensity", 0.0)

	// Color Grading
	p.compositeShader.SetUniform1f("u_Contrast", 1.05)
	p.compositeShader.SetUniform1f("u_Saturation", 1.1)
	p.compositeShader.SetUniform3f("u_ColorTint", 1.0, 1.0, 1.0)

	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, p.sceneTex)
	p.compositeShader.SetUniform1i("u_Scene", 0)

	gl.ActiveTexture(gl.TEXTURE1)
	gl.BindTexture(gl.TEXTURE_2D, p.pingTex[toInt(!horizontal)])
	p.compositeShader.SetUniform1i("u_BloomBlur", 1)

	gl.DrawArrays(gl.TRIANGLES, 0, 6)
	gl.BindVertexArray(0)
}

func (p *PostFX) Delete() {
	gl.DeleteFramebuffers(1, &p.sceneFBO)
	gl.DeleteTextures(1, &p.sceneTex)
	gl.DeleteFramebuffers(2, &p.pingFBO[0])
	gl.DeleteTextures(2, &p.pingTex[0])

	p.extractShader.Delete()
	p.blurShader.Delete()
	p.compositeShader.Delete()

	gl.DeleteVertexArrays(1, &p.quadVAO)
	gl.DeleteBuffers(1, &p.quadVBO)
}

func toInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func toInt32(b bool) int32 {
	if b {
		return 1
	}
	return 0
}

func createScreenQuad() (uint32, uint32) {
	quadVertices := []float32{
		// Pos        // UV
		-1.0,  1.0,  0.0, 1.0,
		-1.0, -1.0,  0.0, 0.0,
		 1.0, -1.0,  1.0, 0.0,

		-1.0,  1.0,  0.0, 1.0,
		 1.0, -1.0,  1.0, 0.0,
		 1.0,  1.0,  1.0, 1.0,
	}

	var vao, vbo uint32
	gl.GenVertexArrays(1, &vao)
	gl.GenBuffers(1, &vbo)

	gl.BindVertexArray(vao)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(quadVertices)*4, gl.Ptr(quadVertices), gl.STATIC_DRAW)

	gl.EnableVertexAttribArray(0)
	gl.VertexAttribPointer(0, 2, gl.FLOAT, false, 4*4, gl.PtrOffset(0))
	gl.EnableVertexAttribArray(1)
	gl.VertexAttribPointer(1, 2, gl.FLOAT, false, 4*4, gl.PtrOffset(2*4))

	gl.BindVertexArray(0)
	return vao, vbo
}

// Shader sources
const vertexShaderSource = `
#version 330 core
layout (location = 0) in vec2 aPos;
layout (location = 1) in vec2 aTexCoords;
out vec2 TexCoords;
void main() {
    TexCoords = aTexCoords;
    gl_Position = vec4(aPos, 0.0, 1.0);
}
` + "\x00"

const extractShaderSource = `
#version 330 core
out vec4 FragColor;
in vec2 TexCoords;
uniform sampler2D u_Scene;
uniform float u_Threshold;
void main() {
    vec4 color = texture(u_Scene, TexCoords);
    float brightness = dot(color.rgb, vec3(0.2126, 0.7152, 0.0722));
    if (brightness > u_Threshold) {
        FragColor = vec4(color.rgb, 1.0);
    } else {
        FragColor = vec4(0.0, 0.0, 0.0, 1.0);
    }
}
` + "\x00"

const blurShaderSource = `
#version 330 core
out vec4 FragColor;
in vec2 TexCoords;
uniform sampler2D u_Image;
uniform bool u_Horizontal;
uniform float u_Weight[5] = float[] (0.227027, 0.1945946, 0.1216216, 0.054054, 0.0162162);
void main() {
    vec2 tex_offset = 1.0 / textureSize(u_Image, 0);
    vec3 result = texture(u_Image, TexCoords).rgb * u_Weight[0];
    if (u_Horizontal) {
        for (int i = 1; i < 5; ++i) {
            result += texture(u_Image, TexCoords + vec2(tex_offset.x * i, 0.0)).rgb * u_Weight[i];
            result += texture(u_Image, TexCoords - vec2(tex_offset.x * i, 0.0)).rgb * u_Weight[i];
        }
    } else {
        for (int i = 1; i < 5; ++i) {
            result += texture(u_Image, TexCoords + vec2(0.0, tex_offset.y * i)).rgb * u_Weight[i];
            result += texture(u_Image, TexCoords - vec2(0.0, tex_offset.y * i)).rgb * u_Weight[i];
        }
    }
    FragColor = vec4(result, 1.0);
}
` + "\x00"

const compositeShaderSource = `
#version 330 core
out vec4 FragColor;
in vec2 TexCoords;

uniform sampler2D u_Scene;
uniform sampler2D u_BloomBlur;
uniform float u_Time;

uniform bool u_BloomEnabled;
uniform float u_BloomIntensity;

uniform bool u_VignetteEnabled;
uniform float u_VignetteIntensity;
uniform float u_VignetteRoundness;

uniform bool u_GrainEnabled;
uniform float u_GrainIntensity;

uniform float u_Contrast;
uniform float u_Saturation;
uniform vec3 u_ColorTint;

float noise(vec2 co) {
    return fract(sin(dot(co.xy ,vec2(12.9898,78.233))) * 43758.5453);
}

void main() {
    vec3 sceneColor = texture(u_Scene, TexCoords).rgb;
    
    if (u_BloomEnabled) {
        vec3 bloomColor = texture(u_BloomBlur, TexCoords).rgb;
        sceneColor += bloomColor * u_BloomIntensity;
    }
    
    sceneColor = (sceneColor - 0.5) * u_Contrast + 0.5;
    
    vec3 luma = vec3(dot(sceneColor, vec3(0.2126, 0.7152, 0.0722)));
    sceneColor = mix(luma, sceneColor, u_Saturation);
    
    sceneColor *= u_ColorTint;
    
    if (u_GrainEnabled) {
        float n = (noise(TexCoords + vec2(u_Time * 0.01, u_Time * 0.02)) - 0.5) * 2.0;
        sceneColor += vec3(n * u_GrainIntensity);
    }
    
    if (u_VignetteEnabled) {
        vec2 uv = TexCoords - 0.5;
        float dist = length(uv);
        float vignette = smoothstep(u_VignetteRoundness, u_VignetteRoundness - u_VignetteIntensity, dist);
        sceneColor *= vignette;
    }
    
    FragColor = vec4(sceneColor, 1.0);
}
` + "\x00"
