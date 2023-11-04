from typing import Any

def copyDict(source: Any, target: Any) -> None:
	for key, value in source.items():
		if isinstance(value, dict):
			copyDict(value, target.setdefault(key, {}))
		else:
			target[key] = value
