from typing import Any, Callable
import logging

logger = logging.getLogger("discord-rich-presence-plex")
logger.setLevel(logging.INFO)
handler = logging.StreamHandler()
handler.setFormatter(logging.Formatter("[%(asctime)s] [%(levelname)s] %(message)s", datefmt = "%d-%m-%Y %I:%M:%S %p"))
logger.addHandler(handler)

class LoggerWithPrefix:

	def __init__(self, prefix: str) -> None:
		self.prefix = prefix
		self.info = self._wrapLoggerFunc(logger.info)
		self.warning = self._wrapLoggerFunc(logger.warning)
		self.error = self._wrapLoggerFunc(logger.error)
		self.exception = self._wrapLoggerFunc(logger.exception)
		self.debug = self._wrapLoggerFunc(logger.debug)

	def _wrapLoggerFunc(self, func: Callable[..., None]) -> Callable[..., None]:
		def wrappedFunc(obj: Any, *args: Any, **kwargs: Any) -> None:
			func(self.prefix + str(obj), *args, **kwargs)
		return wrappedFunc
