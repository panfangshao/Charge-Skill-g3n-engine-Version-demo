// Package meshes builds the primitive geometry the Godot scene uses.
//
// g3n ships Box and Sphere but no capsule, and its sphere does not carry
// Godot's tessellation. Generating all three here keeps the silhouettes and
// the smooth-normal shading identical to BoxMesh/CapsuleMesh/SphereMesh, and
// costs a few hundred vertices.
package meshes

import (
	"math"

	"github.com/g3n/engine/geometry"
	"github.com/g3n/engine/gls"
	"github.com/g3n/engine/math32"
)

const tau = 2 * math.Pi

// builder accumulates a triangle-indexed mesh.
type builder struct {
	positions *math32.ArrayF32
	normals   *math32.ArrayF32
	uvs       *math32.ArrayF32
	indices   *math32.ArrayU32
	count     uint32
}

func newBuilder() *builder {
	p := math32.NewArrayF32(0, 0)
	n := math32.NewArrayF32(0, 0)
	u := math32.NewArrayF32(0, 0)
	i := math32.NewArrayU32(0, 0)
	return &builder{positions: &p, normals: &n, uvs: &u, indices: &i}
}

func (b *builder) add(px, py, pz, nx, ny, nz, u, v float32) uint32 {
	b.positions.Append(px, py, pz)
	b.normals.Append(nx, ny, nz)
	b.uvs.Append(u, v)
	idx := b.count
	b.count++
	return idx
}

func (b *builder) quad(a, bb, c, d uint32) {
	b.indices.Append(a, bb, c)
	b.indices.Append(a, c, d)
}

// geometry uploads the accumulated buffers. g3n derives the bounding box and
// sphere from the position VBO on demand, so there is nothing to set here.
func (b *builder) geometry() *geometry.Geometry {
	g := geometry.NewGeometry()
	g.SetIndices(*b.indices)
	g.AddVBO(gls.NewVBO(*b.positions).AddAttrib(gls.VertexPosition))
	g.AddVBO(gls.NewVBO(*b.normals).AddAttrib(gls.VertexNormal))
	g.AddVBO(gls.NewVBO(*b.uvs).AddAttrib(gls.VertexTexcoord))
	return g
}

// stitchRings joins two consecutive rings of segments+1 vertices each. The seam
// vertex is duplicated in every ring so the normals stay smooth all the way
// around.
func (b *builder) stitchRings(previous, current uint32, segments int) {
	for i := 0; i < segments; i++ {
		i32 := uint32(i)
		b.quad(previous+i32, current+i32, current+i32+1, previous+i32+1)
	}
}

// NewBox is Godot's BoxMesh: flat-shaded, centred on the origin.
func NewBox(size math32.Vector3) *geometry.Geometry {
	half := math32.Vector3{X: size.X / 2, Y: size.Y / 2, Z: size.Z / 2}
	axes := []math32.Vector3{
		{X: 1}, {X: -1}, {Y: 1}, {Y: -1}, {Z: 1}, {Z: -1},
	}

	b := newBuilder()
	for _, normal := range axes {
		// Build a tangent frame for this face, then emit its four corners.
		reference := math32.Vector3{Y: 1}
		if math32.Abs(normal.Y) > 0.5 {
			reference = math32.Vector3{Z: 1}
		}
		tangent := *reference.Clone().Cross(&normal)
		tangent.Normalize()
		bitangent := *normal.Clone().Cross(&tangent)

		centre := math32.Vector3{X: normal.X * half.X, Y: normal.Y * half.Y, Z: normal.Z * half.Z}
		u := math32.Vector3{X: tangent.X * half.X, Y: tangent.Y * half.Y, Z: tangent.Z * half.Z}
		v := math32.Vector3{X: bitangent.X * half.X, Y: bitangent.Y * half.Y, Z: bitangent.Z * half.Z}

		corner := func(su, sv float32) uint32 {
			return b.add(
				centre.X+u.X*su+v.X*sv, centre.Y+u.Y*su+v.Y*sv, centre.Z+u.Z*su+v.Z*sv,
				normal.X, normal.Y, normal.Z,
				(su+1)/2, (sv+1)/2)
		}
		i0 := corner(-1, -1)
		i1 := corner(1, -1)
		i2 := corner(1, 1)
		i3 := corner(-1, 1)
		b.quad(i0, i1, i2, i3)
	}

	return b.geometry()
}

// NewCapsule is Godot's CapsuleMesh: height spans the whole capsule, caps
// included.
func NewCapsule(radius, height float32, radialSegments, rings int) *geometry.Geometry {
	// The cylindrical middle is height - 2*radius tall; each cap adds a radius.
	middle := height - 2*radius
	capCentre := middle / 2

	b := newBuilder()
	var previous uint32
	hasPrevious := false

	emitRing := func(y, ringRadius, normalScale, normalY float32) {
		base := b.count
		for i := 0; i <= radialSegments; i++ {
			u := float32(i) / float32(radialSegments)
			x := -math32.Sin(u * tau)
			z := math32.Cos(u * tau)
			n := math32.Vector3{X: x * normalScale, Y: normalY, Z: -z * normalScale}
			n.Normalize()
			b.add(x*ringRadius, y, -z*ringRadius, n.X, n.Y, n.Z, u, 0)
		}
		if hasPrevious {
			b.stitchRings(previous, base, radialSegments)
		}
		previous = base
		hasPrevious = true
	}

	// Top cap: a quarter turn from the pole down to the equator.
	for j := 0; j <= rings+1; j++ {
		v := float32(j) / float32(rings+1)
		w := math32.Sin(0.5 * math32.Pi * v)
		y := math32.Cos(0.5 * math32.Pi * v)
		emitRing(capCentre+radius*y, radius*w, w, y)
	}
	// Cylinder.
	for j := 0; j <= rings+1; j++ {
		v := float32(j) / float32(rings+1)
		emitRing(capCentre-middle*v, radius, 1, 0)
	}
	// Bottom cap.
	for j := 0; j <= rings+1; j++ {
		v := float32(j) / float32(rings+1)
		w := math32.Cos(0.5 * math32.Pi * v)
		y := math32.Sin(0.5 * math32.Pi * v)
		emitRing(-capCentre-radius*y, radius*w, w, -y)
	}

	return b.geometry()
}

// NewSphere is Godot's SphereMesh, which allows a height different from the
// diameter.
func NewSphere(radius, height float32, radialSegments, rings int) *geometry.Geometry {
	halfHeight := height / 2

	b := newBuilder()
	var previous uint32

	for j := 0; j <= rings+1; j++ {
		v := float32(j) / float32(rings+1)
		w := math32.Sin(math32.Pi * v)
		y := math32.Cos(math32.Pi * v)

		base := b.count
		for i := 0; i <= radialSegments; i++ {
			u := float32(i) / float32(radialSegments)
			x := -math32.Sin(u * tau)
			z := math32.Cos(u * tau)
			n := math32.Vector3{X: x * w, Y: y, Z: -z * w}
			n.Normalize()
			b.add(x*radius*w, halfHeight*y, -z*radius*w, n.X, n.Y, n.Z, u, v)
		}
		if j > 0 {
			b.stitchRings(previous, base, radialSegments)
		}
		previous = base
	}

	return b.geometry()
}
