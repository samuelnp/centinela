package importgraph

// pyASTScript is the embedded import-graph walker run via `python3 -c`. It
// discovers internal modules (by .py path), parses each via the ast module, and
// emits the shared JSON contract: project-relative package paths plus their
// intra-project import edges. External/stdlib imports are dropped (an import is
// internal only when it resolves to a discovered module path).
const pyASTScript = `
import ast, json, os, sys
root = sys.argv[1] if len(sys.argv) > 1 else "."
skip = {"venv", ".venv", "__pycache__", "node_modules", "build", "dist"}
files = []
for dp, dirs, fns in os.walk(root):
    dirs[:] = [d for d in dirs if not d.startswith(".") and d not in skip]
    files += [os.path.join(dp, f) for f in fns if f.endswith(".py")]
def dotted(f):
    rel = os.path.relpath(f, root)[:-3].split(os.sep)
    if rel and rel[-1] == "__init__":
        rel = rel[:-1]
    return ".".join(rel)
known = {dotted(f) for f in files}
def match(name):
    if name in known:
        return name
    top = name.split(".")[0]
    return top if top in known else None
pkgs = {}
for f in files:
    try:
        tree = ast.parse(open(f, encoding="utf-8", errors="replace").read())
    except SyntaxError:
        continue
    cur = dotted(f).replace(".", "/")
    imps = set()
    for n in ast.walk(tree):
        if isinstance(n, ast.Import):
            for a in n.names:
                m = match(a.name)
                if m:
                    imps.add(m.replace(".", "/"))
        elif isinstance(n, ast.ImportFrom) and n.module and not n.level:
            m = match(n.module)
            if m:
                imps.add(m.replace(".", "/"))
    pkgs.setdefault(cur, set()).update(i for i in imps if i != cur)
out = {"module": os.path.basename(os.path.abspath(root)),
       "pkgs": [{"path": k, "imports": sorted(v)} for k, v in sorted(pkgs.items())]}
print(json.dumps(out))
`
