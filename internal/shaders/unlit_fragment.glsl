// Godot's SHADING_MODE_UNSHADED, for the grid strips and the effect puffs.
//
// Godot still runs those through its tonemapper before they hit the frame
// buffer, so the colour arrives here already baked by colors.Unshaded and this
// shader only has to write it.

precision highp float;

out vec4 FragColor;

uniform vec3 MatParams[2];
#define MatColor MatParams[0]
#define MatAlpha MatParams[1].x

void main() {
    FragColor = vec4(MatColor, MatAlpha);
}
