def formatSeconds(seconds: int, joiner: str = "") -> str:
	seconds = round(seconds)
	timeValues = {"h": seconds // 3600, "m": seconds // 60 % 60, "s": seconds % 60}
	if not joiner:
		return "".join(str(v) + k for k, v in timeValues.items() if v > 0)
	if timeValues["h"] == 0:
		del timeValues["h"]
	return joiner.join(str(v).rjust(2, "0") for v in timeValues.values())
