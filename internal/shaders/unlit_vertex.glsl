// Vertex half of the unlit material. See unlit_fragment.glsl.

#include <attributes>

uniform mat4 MVP;

void main() {
    gl_Position = MVP * vec4(VertexPosition, 1.0);
}
