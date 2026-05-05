package renderer

import (
	"fmt"
	"strings"

	"github.com/go-gl/gl/v3.3-core/gl"
)

type Shader struct {
	ID uint32
}

func CompileShader(source string, shaderType uint32) (uint32, error) {
	shader := gl.CreateShader(shaderType)

	csources, free := gl.Strs(source + "\x00")
	gl.ShaderSource(shader, 1, csources, nil)
	free()
	gl.CompileShader(shader)

	var status int32
	gl.GetShaderiv(shader, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetShaderiv(shader, gl.INFO_LOG_LENGTH, &logLength)

		logText := strings.Repeat("\x00", int(logLength+1))
		gl.GetShaderInfoLog(shader, logLength, nil, gl.Str(logText))

		return 0, fmt.Errorf("failed to compile shader %d: %s", shaderType, logText)
	}

	return shader, nil
}

func NewShader(vertexSource, fragmentSource string) (*Shader, error) {
	vertexShader, err := CompileShader(vertexSource, gl.VERTEX_SHADER)
	if err != nil {
		return nil, err
	}
	defer gl.DeleteShader(vertexShader)

	fragmentShader, err := CompileShader(fragmentSource, gl.FRAGMENT_SHADER)
	if err != nil {
		return nil, err
	}
	defer gl.DeleteShader(fragmentShader)

	program := gl.CreateProgram()

	gl.AttachShader(program, vertexShader)
	gl.AttachShader(program, fragmentShader)
	gl.LinkProgram(program)

	var status int32
	gl.GetProgramiv(program, gl.LINK_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetProgramiv(program, gl.INFO_LOG_LENGTH, &logLength)

		logText := strings.Repeat("\x00", int(logLength+1))
		gl.GetProgramInfoLog(program, logLength, nil, gl.Str(logText))

		return nil, fmt.Errorf("failed to link shader program: %s", logText)
	}

	return &Shader{ID: program}, nil
}

func (s *Shader) Use() {
	gl.UseProgram(s.ID)
}

func (s *Shader) SetUniformMat4(name string, matrix *[16]float32) {
	location := gl.GetUniformLocation(s.ID, gl.Str(name+"\x00"))
	gl.UniformMatrix4fv(location, 1, false, &matrix[0])
}

func (s *Shader) SetUniform1i(name string, value int32) {
	location := gl.GetUniformLocation(s.ID, gl.Str(name+"\x00"))
	gl.Uniform1i(location, value)
}

func (s *Shader) SetUniform1f(name string, value float32) {
	location := gl.GetUniformLocation(s.ID, gl.Str(name+"\x00"))
	gl.Uniform1f(location, value)
}

func (s *Shader) SetUniform2f(name string, x, y float32) {
	location := gl.GetUniformLocation(s.ID, gl.Str(name+"\x00"))
	gl.Uniform2f(location, x, y)
}

func (s *Shader) SetUniform3f(name string, x, y, z float32) {
	location := gl.GetUniformLocation(s.ID, gl.Str(name+"\x00"))
	gl.Uniform3f(location, x, y, z)
}

func (s *Shader) Delete() {
	gl.DeleteProgram(s.ID)
}
