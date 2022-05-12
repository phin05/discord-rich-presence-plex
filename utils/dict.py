from typing import Any

def merge(source: Any, target: Any) -> None:
	for key, value in source.items():
		if isinstance(value, dict):
			merge(value, target.setdefault(key, {}))
		else:
			target[key] = source[key]
