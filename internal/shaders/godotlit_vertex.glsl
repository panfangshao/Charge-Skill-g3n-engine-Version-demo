// Vertex half of the Godot gl_compatibility port. See godotlit_fragment.glsl.
//
// g3n's own shaders light in camera space; this one works in world space
// instead, because the sun direction and the two shadow-casting capsules are
// naturally expressed there and would otherwise have to be re-derived every
// frame. ModelMatrix is already uploaded by graphic.Mesh.RenderSetup.

#include <attributes>

uniform mat4 ModelMatrix;
uniform mat4 MVP;

out vec3 WorldPos;
out vec3 WorldNormal;

void main() {
    vec4 world = ModelMatrix * vec4(VertexPosition, 1.0);
    WorldPos = world.xyz;

    // Every object in this scene carries uniform scale and no rotation, so the
    // upper 3x3 of the model matrix is a similarity transform and normalizing
    // it is the same as using the inverse transpose.
    WorldNormal = normalize(mat3(ModelMatrix) * VertexNormal);

    gl_Position = MVP * vec4(VertexPosition, 1.0);
}
