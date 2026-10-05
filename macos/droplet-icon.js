// Renders the drop box app icon (macos/droplet-icon.png): the Phosphor
// "tray-arrow-down" glyph (MIT, https://phosphoricons.com) in white on a
// blue macOS-style tile.
//
//   npm install --no-save @resvg/resvg-js
//   node macos/droplet-icon.js
const fs = require("fs");
const { Resvg } = require("@resvg/resvg-js");
const path = require("path");
const glyphPath = path.join(__dirname, "tray-arrow-down-fill.svg");
const out = path.join(__dirname, "droplet-icon.png");
const [size, top, bot, ink, scale] = ["1024", "#5AB0FF", "#2563EB", "#FFFFFF", "0.66"];
let g = fs.readFileSync(glyphPath, "utf8");
const vb = (g.match(/viewBox="([^"]+)"/) || [, "0 0 24 24"])[1];
const inner = g.replace(/^[\s\S]*?<svg[^>]*>/, "").replace(/<\/svg>\s*$/, "");
const strokeAttrs = /stroke="currentColor"/.test(g)
  ? `fill="none" stroke="${ink}" stroke-width="${(g.match(/stroke-width="([\d.]+)"/) || [, 2])[1]}" stroke-linecap="round" stroke-linejoin="round"`
  : `fill="${ink}"`;
const box = 824 * Number(scale), off = 512 - box / 2;
const svg = `<svg xmlns="http://www.w3.org/2000/svg" width="1024" height="1024" viewBox="0 0 1024 1024">
<defs>
 <linearGradient id="bg" x1="0" y1="0" x2="0" y2="1"><stop offset="0" stop-color="${top}"/><stop offset="1" stop-color="${bot}"/></linearGradient>
 <filter id="sh" x="-20%" y="-20%" width="140%" height="140%"><feDropShadow dx="0" dy="12" stdDeviation="14" flood-color="#000" flood-opacity="0.28"/></filter>
 <filter id="gs"><feDropShadow dx="0" dy="6" stdDeviation="8" flood-color="#000" flood-opacity="0.18"/></filter>
</defs>
<rect x="100" y="100" width="824" height="824" rx="185" fill="url(#bg)" filter="url(#sh)"/>
<svg x="${off}" y="${off + 6}" width="${box}" height="${box}" viewBox="${vb}" filter="url(#gs)"><g ${strokeAttrs}>${inner}</g></svg>
</svg>`;
const png = new Resvg(svg, { fitTo: { mode: "width", value: Number(size) } }).render().asPng();
fs.writeFileSync(out, png);
