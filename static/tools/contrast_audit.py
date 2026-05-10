import re
import os
import sys
from math import pow

CSS_PATHS = [
    'static/css/design-system.css',
    'static/css/theme.css',
    'static/css/components.css',
    'static/css/style.css',
    'static/css/platform.css'
]

VAR_RE = re.compile(r'--([a-zA-Z0-9-_]+)\s*:\s*([^;]+);')
HEX_RE = re.compile(r'#([0-9a-fA-F]{3,8})')


def hex_to_rgb(hexstr):
    h = hexstr.lstrip('#')
    if len(h) in (3,4):
        h = ''.join([c*2 for c in h])
    if len(h) == 6:
        r,g,b = int(h[0:2],16), int(h[2:4],16), int(h[4:6],16)
        a = 255
    elif len(h) == 8:
        r,g,b,a = int(h[0:2],16), int(h[2:4],16), int(h[4:6],16), int(h[6:8],16)
    else:
        return None
    return (r,g,b,a)


def srgb_to_linear(c):
    c = c/255.0
    if c <= 0.03928:
        return c/12.92
    return pow((c+0.055)/1.055, 2.4)


def luminance(rgb):
    r,g,b,_ = rgb
    R = srgb_to_linear(r)
    G = srgb_to_linear(g)
    B = srgb_to_linear(b)
    return 0.2126*R + 0.7152*G + 0.0722*B


def contrast_ratio(rgb1, rgb2):
    L1 = luminance(rgb1)
    L2 = luminance(rgb2)
    hi = max(L1,L2)
    lo = min(L1,L2)
    return (hi + 0.05) / (lo + 0.05)


def extract_vars(content):
    vars = {}
    for m in VAR_RE.finditer(content):
        name = m.group(1).strip()
        val = m.group(2).strip()
        # try to extract hex
        hx = HEX_RE.search(val)
        if hx:
            vars[name] = '#' + hx.group(1)
    return vars


def load_css_vars():
    vars = {}
    for p in CSS_PATHS:
        if not os.path.exists(p):
            continue
        with open(p, 'r', encoding='utf-8') as f:
            content = f.read()
        found = extract_vars(content)
        for k,v in found.items():
            vars[k] = v
    return vars


def find_pairs(vars):
    pairs = []
    # common names
    keys = list(vars.keys())
    for k in keys:
        low = k.lower()
        if 'bg' in low or 'background' in low or 'surface' in low:
            # try to find a text color
            for j in keys:
                if j==k: continue
                lj = j.lower()
                if 'text' in lj or 'fg' in lj or 'color' in lj or 'contrast' in lj:
                    pairs.append((k,j))
    # fallback: any obvious text/bg combos
    if not pairs:
        # pick --background, --bg, --surface as bg; --text, --fg as text
        bg_keys = [x for x in keys if any(y in x.lower() for y in ['bg','background','surface','card','paper'])]
        text_keys = [x for x in keys if any(y in x.lower() for y in ['text','fg','color','on-'])]
        for b in bg_keys:
            for t in text_keys:
                if b!=t:
                    pairs.append((b,t))
    return pairs


def main():
    vars = load_css_vars()
    if not vars:
        print('No CSS variables found in:', CSS_PATHS)
        sys.exit(1)
    print('Found %d CSS color variables (hex):' % len(vars))
    for k,v in vars.items():
        print('  --%s: %s' % (k, v))
    pairs = find_pairs(vars)
    if not pairs:
        print('\nNo sensible bg/text pairs detected. Listing all combos...')
        keys = list(vars.keys())
        for i in range(len(keys)):
            for j in range(i+1, len(keys)):
                pairs.append((keys[i], keys[j]))
    print('\nChecking contrast ratios for candidate pairs:')
    issues = []
    for b,t in pairs:
        hb = vars.get(b)
        ht = vars.get(t)
        if not hb or not ht: continue
        rb = hex_to_rgb(hb)
        rt = hex_to_rgb(ht)
        if not rb or not rt: continue
        ratio = contrast_ratio(rb, rt)
        status = 'PASS' if ratio>=4.5 else ('AA-large' if ratio>=3.0 else 'FAIL')
        print('  %s (bg=%s) vs %s (text=%s): %.2f -> %s' % (b, hb, t, ht, ratio, status))
        if status!='PASS':
            issues.append((b,t,ratio,status))
    if issues:
        print('\nIssues found (%d):' % len(issues))
        for b,t,ratio,status in issues:
            print(' - %s vs %s: %.2f (%s)' % (b,t,ratio,status))
        sys.exit(2)
    print('\nAll checked pairs pass AA contrast for normal text (>=4.5:1).')

if __name__=='__main__':
    main()
