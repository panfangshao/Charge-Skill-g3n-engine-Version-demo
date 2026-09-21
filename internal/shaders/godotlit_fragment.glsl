// Port of Godot 4's gl_compatibility scene shader for the subset this demo
// needs: one directional light, a flat colour ambient term, the Schlick-GGX
// specular lobe, Lambert diffuse and the filmic tonemapper.
//
// Godot's compatibility renderer tonemaps *per fragment* -- there is no HDR
// resolve pass -- and then encodes sRGB on the spot, which is why both live
// here rather than in a post-process. Whatever this shader writes is what lands
// in the frame buffer, exactly as in Godot, and the translucent effect puffs
// and the HUD blend against sRGB values the same way they do there.
//
// Shadows are analytic. g3n has no shadow mapping at all, but this scene has
// exactly two shadow casters and both are upright capsules, so the shadow test
// reduces to the distance between the sun ray and a vertical segment -- a
// closed-form clamped closest-point. That is cheaper than a shadow map and has
// no bias or acne to tune.

precision highp float;

in vec3 WorldPos;
in vec3 WorldNormal;

out vec4 FragColor;

uniform vec3 CameraPosition;

// MatParams packs the Godot StandardMaterial3D settings and the scene lighting:
//   [0] albedo in sRGB, as Godot stores it
//   [1] linear ambient radiance, energy and calibration gain folded in
//   [2] linear directional radiance, energy and calibration gain folded in
//   [3] unit vector pointing *towards* the sun, in world space
//   [4] (metallic, roughness, alpha)
uniform vec3 MatParams[5];
#define MatAlbedo    MatParams[0]
#define MatAmbient   MatParams[1]
#define MatSun       MatParams[2]
#define MatSunDir    MatParams[3]
#define MatMetallic  MatParams[4].x
#define MatRoughness MatParams[4].y
#define MatAlpha     MatParams[4].z

// The two shadow casters:
//   [0], [1] capsule centres in world space
//   [2]      (radius, halfSegment, skipIndex)
// skipIndex is 1 or 2 when the surface being shaded is itself that capsule, so
// a capsule is shaded by N.L alone instead of fighting its own silhouette.
uniform vec3 ShadowParams[3];
#define ShadowRadius  ShadowParams[2].x
#define ShadowHalfLen ShadowParams[2].y
#define ShadowSkip    ShadowParams[2].z

const float M_PI = 3.141592653589793;

// The exact sRGB transfer function. The usual cheap polynomial fit drifts in
// the near-black range this scene is full of.
vec3 SrgbToLinear(vec3 c) {
    return mix(c / 12.92, pow((c + 0.055) / 1.055, vec3(2.4)), step(vec3(0.04045), c));
}

vec3 LinearToSrgb(vec3 c) {
    c = max(c, vec3(0.0));
    return mix(c * 12.92, 1.055 * pow(c, vec3(1.0 / 2.4)) - 0.055, step(vec3(0.0031308), c));
}

float D_GGX(float cosThetaM, float alpha) {
    float alpha2 = alpha * alpha;
    float d = 1.0 + (alpha2 - 1.0) * cosThetaM * cosThetaM;
    return alpha2 / (M_PI * d * d);
}

float V_GGX(float ndotl, float ndotv, float alpha) {
    return 0.5 / mix(2.0 * ndotl * ndotv, ndotl + ndotv, alpha);
}

float SchlickFresnel(float u) {
    float m = clamp(1.0 - u, 0.0, 1.0);
    float m2 = m * m;
    return m2 * m2 * m;
}

// Godot: tonemap_filmic() with exposure_bias 2.0 folded into A and B, white 1.
vec3 TonemapFilmic(vec3 color) {
    const float a = 0.88; // 0.22 * bias * bias
    const float b = 0.60; // 0.30 * bias
    const float c = 0.10;
    const float d = 0.20;
    const float e = 0.01;
    const float f = 0.30;
    const float white = 1.0;

    vec3 mapped = ((color * (a * color + c * b) + d * e) / (color * (a * color + b) + d * f)) - e / f;
    float whiteMapped = ((white * (a * white + c * b) + d * e) / (white * (a * white + b) + d * f)) - e / f;
    return mapped / whiteMapped;
}

// How much sun reaches `p` past one upright capsule.
//
// Closest approach between the sun ray p + l*t (t >= 0) and the capsule's
// spine, the vertical segment of half-length `halfLen` through `centre`. Both
// parameters are clamped to their valid ranges, which is the standard
// segment-segment closest-point solve; the sun is never near vertical here, so
// the denominator stays well away from zero.
float ShadowFromCapsule(vec3 p, vec3 centre, float halfLen, float radius) {
    vec3 l = MatSunDir;
    vec3 r = p - centre;

    float ly = l.y;
    float denom = 1.0 - ly * ly;
    float t = 0.0;
    if (denom > 1e-4) {
        t = (ly * r.y - dot(l, r)) / denom;
    }
    t = max(t, 0.0);

    float s = clamp(ly * t + r.y, -halfLen, halfLen);
    // Re-solve t for the clamped point on the spine.
    vec3 spine = centre + vec3(0.0, s, 0.0);
    t = max(dot(l, spine - p), 0.0);

    float dist = length(p + l * t - spine);
    // Soften over a couple of centimetres so the edge is not a hard jaggy.
    return smoothstep(radius - 0.03, radius + 0.03, dist);
}

void main() {
    vec3 albedo = SrgbToLinear(MatAlbedo);
    float metallic = MatMetallic;
    float roughness = MatRoughness;

    vec3 n = normalize(WorldNormal);
    // Godot derives the view vector from the fragment position even under an
    // orthographic projection, which is what gives its ground that faint
    // horizontal specular gradient.
    vec3 v = normalize(CameraPosition - WorldPos);
    vec3 l = normalize(MatSunDir);
    vec3 h = normalize(l + v);

    float ndotl = max(dot(n, l), 0.0);
    float ndotv = max(dot(n, v), 1e-4);
    float ndoth = clamp(dot(n, h), 0.0, 1.0);
    float ldoth = clamp(dot(l, h), 0.0, 1.0);

    float shadow = 1.0;
    if (ShadowSkip != 1.0) {
        shadow *= ShadowFromCapsule(WorldPos, ShadowParams[0], ShadowHalfLen, ShadowRadius);
    }
    if (ShadowSkip != 2.0) {
        shadow *= ShadowFromCapsule(WorldPos, ShadowParams[1], ShadowHalfLen, ShadowRadius);
    }

    vec3 lightColor = MatSun * shadow;

    vec3 f0 = mix(vec3(0.04), albedo, metallic);
    float f90 = clamp(50.0 * f0.g, 0.0, 1.0);
    float alphaGgx = roughness * roughness;

    vec3 diffuseLight = albedo * (1.0 - metallic) * (MatAmbient + lightColor * ndotl);

    vec3 fresnel = f0 + (f90 - f0) * SchlickFresnel(ldoth);
    vec3 specularLight = lightColor * ndotl * D_GGX(ndoth, alphaGgx) * V_GGX(ndotl, ndotv, alphaGgx) * fresnel;

    vec3 color = diffuseLight + specularLight;
    FragColor = vec4(LinearToSrgb(TonemapFilmic(color)), MatAlpha);
}
