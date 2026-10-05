# Renders the drop box app icon (macos/droplet-icon.png): a line-art open
# box with an arrow dropping in, on a light macOS-style tile.
#
#   python3 macos/droplet-icon.py macos/droplet-icon.png   (needs Pillow)
from PIL import Image, ImageDraw, ImageFilter
import sys

K = 4                                   # supersampling
S = 1024 * K
W = 44                                  # stroke width
INK = (43, 47, 54)
ACCENT = (18, 160, 130)


DY = 18                                 # optical centering on the tile


def P(x, y):
    return (x * K, y * K)


def Q(x, y):
    return P(x, y + DY)


img = Image.new("RGBA", (S, S), (0, 0, 0, 0))

shadow = Image.new("RGBA", (S, S), (0, 0, 0, 0))
ImageDraw.Draw(shadow).rounded_rectangle([*P(100, 112), *P(924, 936)], radius=185 * K, fill=(0, 0, 0, 70))
img.alpha_composite(shadow.filter(ImageFilter.GaussianBlur(12 * K)))

tile = Image.new("RGBA", (S, S))
td = ImageDraw.Draw(tile)
top, bot = (255, 255, 255), (232, 235, 240)
for y in range(S):
    t = y / S
    td.line([(0, y), (S, y)], fill=tuple(int(a + (b - a) * t) for a, b in zip(top, bot)) + (255,))
mask = Image.new("L", (S, S), 0)
ImageDraw.Draw(mask).rounded_rectangle([*P(100, 100), *P(924, 924)], radius=185 * K, fill=255)
img.paste(tile, (0, 0), mask)

d = ImageDraw.Draw(img)


def stroke(points, color):
    pts = [Q(*pt) for pt in points]
    d.line(pts, fill=color, width=W * K, joint="curve")
    r = W * K / 2
    for x, y in pts:
        d.ellipse([x - r, y - r, x + r, y + r], fill=color)


# box: body with a gap in the rim for the arrow, detached flaps folded out
stroke([(440, 560), (322, 560), (322, 770), (702, 770), (702, 560), (584, 560)], INK)
stroke([(284, 516), (226, 454)], INK)
stroke([(740, 516), (798, 454)], INK)

# arrow dropping through the gap
stroke([(512, 200), (512, 676)], ACCENT)
stroke([(444, 608), (512, 676), (580, 608)], ACCENT)

img.resize((1024, 1024), Image.LANCZOS).save(sys.argv[1])
