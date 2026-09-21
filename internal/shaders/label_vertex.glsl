// Vertex half of the Label3D material. graphic.Sprite uploads MVP and nothing
// else, and that is all this needs.

#include <attributes>

uniform mat4 MVP;

out vec2 FragTexcoord;

void main() {
    FragTexcoord = VertexTexcoord;
    gl_Position = MVP * vec4(VertexPosition, 1.0);
}
