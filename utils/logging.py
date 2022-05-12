from typing import Any
import logging

logger = logging.getLogger("discord-rich-presence-plex")
logger.setLevel(logging.INFO)
handler = logging.StreamHandler()
handler.setFormatter(logging.Formatter("[%(asctime)s] [%(levelname)s] %(message)s", datefmt = "%d-%m-%Y %I:%M:%S %p"))
logger.addHandler(handler)

class LoggerWithPrefix:

	def __init__(self, prefix: str) -> None:
		self.prefix = prefix

	def info(self, obj: Any, *args: Any, **kwargs: Any) -> None:
		logger.info(self.prefix + str(obj), *args, **kwargs)

	def warning(self, obj: Any, *args: Any, **kwargs: Any) -> None:
		logger.warning(self.prefix + str(obj), *args, **kwargs)

	def error(self, obj: Any, *args: Any, **kwargs: Any) -> None:
		logger.error(self.prefix + str(obj), *args, **kwargs)

	def exception(self, obj: Any, *args: Any, **kwargs: Any) -> None:
		logger.exception(self.prefix + str(obj), *args, **kwargs)

	def debug(self, obj: Any, *args: Any, **kwargs: Any) -> None:
		logger.debug(self.prefix + str(obj), *args, **kwargs)
