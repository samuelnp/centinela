package roadmap

func hasCycle(deps map[string][]string) bool {
	state := map[string]int{}
	var dfs func(string) bool
	dfs = func(n string) bool {
		if state[n] == 1 {
			return true
		}
		if state[n] == 2 {
			return false
		}
		state[n] = 1
		for _, d := range deps[n] {
			if dfs(d) {
				return true
			}
		}
		state[n] = 2
		return false
	}
	for n := range deps {
		if dfs(n) {
			return true
		}
	}
	return false
}
