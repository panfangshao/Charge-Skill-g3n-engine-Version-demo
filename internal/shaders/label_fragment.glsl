// Fragment half of the Label3D material: an unlit, alpha-blended text texture.
// The colours were baked through Godot's tonemapper when the text was
// rasterised, so nothing happens to them here.

precision highp float;

in vec2 FragTexcoord;

out vec4 FragColor;

uniform sampler2D MatTexture[1];
uniform vec2 MatTexinfo[3];

void main() {
    // g3n rasterises text top-down; the quad's V axis runs the other way.
    vec2 uv = vec2(FragTexcoord.x, 1.0 - FragTexcoord.y);
    FragColor = texture(MatTexture[0], uv);
}
