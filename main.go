package main

import (
	"math"
	"math/rand"
	"runtime"
	"time"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"
)

const (
	winWidth  = 1280
	winHeight = 720

	nParticles = 2000 // Reduced slightly for CPU performance
	fixedDT    = 1.0 / 60.0

	coulombK  = 200.0
	magneticK = 0.5
	pressureK = 5.0
	dragK     = 0.05
)

type Particle struct {
	Pos     mgl32.Vec3
	PrevPos mgl32.Vec3
	Vel     mgl32.Vec3
	Col     mgl32.Vec3
	Charge  float32
}

var (
	particles []Particle
	prog      uint32
	vao, vbo  uint32
	timeAcc   float32
)

func init() {
	runtime.LockOSThread()
}

var vertexShader = `
#version 410 core
layout(location = 0) in vec3 inPos;
layout(location = 1) in vec3 inCol;
uniform mat4 uVP;
out vec3 vColor;
void main() {
    gl_Position = uVP * vec4(inPos, 1.0);
    gl_PointSize = 2.0;
    vColor = inCol;
}
` + "\x00"

var fragmentShader = `
#version 410 core
in vec3 vColor;
out vec4 fragColor;
void main() {
    fragColor = vec4(vColor, 0.8);
}
` + "\x00"

func main() {
	if err := glfw.Init(); err != nil {
		panic(err)
	}
	defer glfw.Terminate()

	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, gl.TRUE)

	window, err := glfw.CreateWindow(winWidth, winHeight, "Plasma Field Sim", nil, nil)
	if err != nil {
		panic(err)
	}
	window.MakeContextCurrent()

	if err := gl.Init(); err != nil {
		panic(err)
	}

	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE)
	gl.Enable(gl.PROGRAM_POINT_SIZE)

	var errProg error
	prog, errProg = newProgram(vertexShader, fragmentShader)
	if errProg != nil {
		panic(errProg)
	}

	setupParticles()
	setupBuffers()

	prev := time.Now()
	var accumulator float64

	for !window.ShouldClose() {
		now := time.Now()
		elapsed := now.Sub(prev).Seconds()
		prev = now
		if elapsed > 0.1 {
			elapsed = 0.1
		}
		accumulator += elapsed

		for accumulator >= fixedDT {
			stepPhysics(fixedDT)
			accumulator -= fixedDT
		}

		render(window)
		window.SwapBuffers()
		glfw.PollEvents()
	}
}

func setupParticles() {
	particles = make([]Particle, nParticles)
	for i := range particles {
		pos := mgl32.Vec3{
			(rand.Float32() - 0.5) * 100,
			(rand.Float32() - 0.5) * 100,
			(rand.Float32() - 0.5) * 100,
		}
		charge := float32(1.0)
		color := mgl32.Vec3{1, 0.4, 0.2}
		if rand.Float32() > 0.5 {
			charge = -1.0
			color = mgl32.Vec3{0.2, 0.6, 1.0}
		}
		particles[i] = Particle{Pos: pos, PrevPos: pos, Col: color, Charge: charge}
	}
}

func stepPhysics(dt float64) {
	dt32 := float32(dt)
	for i := 0; i < nParticles; i++ {
		p := &particles[i]
		acc := mgl32.Vec3{0, 0, 0}

		// N-Body Coulomb Force (simplified for performance)
		for j := 0; j < nParticles; j += 2 { // Step by 2 to save CPU cycles
			if i == j {
				continue
			}
			other := particles[j]
			dir := p.Pos.Sub(other.Pos)
			distSq := dir.LenSqr() + 10.0
			if distSq < 5000 {
				acc = acc.Add(dir.Normalize().Mul((coulombK * p.Charge * other.Charge) / distSq))
			}
		}

		// Lorentz-ish Force & Turbulence
		B := mgl32.Vec3{0, 1, 0}
		acc = acc.Add(p.Vel.Cross(B).Mul(magneticK * p.Charge))
		acc = acc.Add(curlNoise(p.Pos.Mul(0.01), timeAcc).Mul(15.0))

		// Gravity/Pressure to center
		acc = acc.Add(p.Pos.Mul(-pressureK * 0.01))
		acc = acc.Add(p.Vel.Mul(-dragK))

		// Verlet Integration
		nextPos := p.Pos.Mul(2).Sub(p.PrevPos).Add(acc.Mul(dt32 * dt32))
		p.Vel = nextPos.Sub(p.Pos).Mul(1.0 / dt32)
		p.PrevPos = p.Pos
		p.Pos = nextPos
	}
	timeAcc += dt32
}

func render(w *glfw.Window) {
	gl.ClearColor(0, 0, 0.05, 1)
	gl.Clear(gl.COLOR_BUFFER_BIT)

	gl.UseProgram(prog)

	// --- Perspective Matrix ---
	// Field of View, Aspect, Near, Far
	view := mgl32.LookAtV(mgl32.Vec3{0, 0, 300}, mgl32.Vec3{0, 0, 0}, mgl32.Vec3{0, 1, 0})
	proj := mgl32.Perspective(mgl32.DegToRad(45), float32(winWidth)/float32(winHeight), 0.1, 1000.0)
	vp := proj.Mul4(view)

	uVPLoc := gl.GetUniformLocation(prog, gl.Str("uVP\x00"))
	gl.UniformMatrix4fv(uVPLoc, 1, false, &vp[0])

	// Update VBO
	data := make([]float32, 0, nParticles*6)
	for _, p := range particles {
		data = append(data, p.Pos.X(), p.Pos.Y(), p.Pos.Z(), p.Col.X(), p.Col.Y(), p.Col.Z())
	}
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.BufferSubData(gl.ARRAY_BUFFER, 0, len(data)*4, gl.Ptr(data))

	gl.BindVertexArray(vao)
	gl.DrawArrays(gl.POINTS, 0, int32(nParticles))
}

// Boilerplate Helpers
func curlNoise(p mgl32.Vec3, t float32) mgl32.Vec3 {
	n := func(v mgl32.Vec3) float32 {
		return float32(math.Sin(float64(v.X()+t)) * math.Cos(float64(v.Y())) * math.Sin(float64(v.Z()-t)))
	}
	e := float32(0.1)
	return mgl32.Vec3{
		n(p.Add(mgl32.Vec3{0, e, 0})) - n(p.Sub(mgl32.Vec3{0, e, 0})),
		n(p.Add(mgl32.Vec3{0, 0, e})) - n(p.Sub(mgl32.Vec3{0, 0, e})),
		n(p.Add(mgl32.Vec3{e, 0, 0})) - n(p.Sub(mgl32.Vec3{e, 0, 0})),
	}.Mul(1.0 / (2 * e))
}

func setupBuffers() {
	gl.GenVertexArrays(1, &vao)
	gl.BindVertexArray(vao)
	gl.GenBuffers(1, &vbo)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.BufferData(gl.ARRAY_BUFFER, nParticles*6*4, nil, gl.DYNAMIC_DRAW)
	gl.EnableVertexAttribArray(0)
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 6*4, gl.PtrOffset(0))
	gl.EnableVertexAttribArray(1)
	gl.VertexAttribPointer(1, 3, gl.FLOAT, false, 6*4, gl.PtrOffset(3*4))
}

func compileShader(source string, shaderType uint32) (uint32, error) {
	shader := gl.CreateShader(shaderType)
	csources, free := gl.Strs(source)
	gl.ShaderSource(shader, 1, csources, nil)
	free()
	gl.CompileShader(shader)
	return shader, nil
}

func newProgram(vertSrc, fragSrc string) (uint32, error) {
	v, _ := compileShader(vertSrc, gl.VERTEX_SHADER)
	f, _ := compileShader(fragSrc, gl.FRAGMENT_SHADER)
	p := gl.CreateProgram()
	gl.AttachShader(p, v)
	gl.AttachShader(p, f)
	gl.LinkProgram(p)
	return p, nil
}
