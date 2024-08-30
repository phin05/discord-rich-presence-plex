from typing import Optional
import re

def formatSeconds(seconds: int | float, joiner: Optional[str] = None) -> str:
	seconds = round(seconds)
	timeValues = { "h": seconds // 3600, "m": seconds // 60 % 60, "s": seconds % 60 }
	if not joiner:
		return "".join(str(v) + k for k, v in timeValues.items() if v > 0)
	if timeValues["h"] == 0:
		del timeValues["h"]
	return joiner.join(str(v).rjust(2, "0") for v in timeValues.values())

def truncate(text: str, maxLength: int) -> str:
	if len(text) > maxLength:
		text = text[:maxLength-3] + "..."
	return text

def stripNonAscii(text: str) -> str:
	return re.sub(r"[^\x00-\x7f]", "", text)
