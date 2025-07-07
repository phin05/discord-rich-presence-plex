from app import constants
from typing import Any, Callable
import logging

logger = logging.getLogger(constants.name)
logger.setLevel(logging.INFO)
formatter = logging.Formatter("[%(asctime)s] [%(levelname)s] %(message)s", datefmt = "%d-%m-%Y %I:%M:%S %p")
streamHandler = logging.StreamHandler()
streamHandler.setFormatter(formatter)
logger.addHandler(streamHandler)

info = logger.info
warning = logger.warning
error = logger.error
exception = logger.exception
debug = logger.debug

class LoggerWithPrefix:

	def __init__(self, prefix: str):
		self.prefix = prefix
		self.info = self.wrapLoggerFunc(logger.info)
		self.warning = self.wrapLoggerFunc(logger.warning)
		self.error = self.wrapLoggerFunc(logger.error)
		self.exception = self.wrapLoggerFunc(logger.exception)
		self.debug = self.wrapLoggerFunc(logger.debug)

	def wrapLoggerFunc(self, func: Callable[..., None]) -> Callable[..., None]:
		def wrappedFunc(obj: Any, *args: Any, **kwargs: Any) -> None:
			func(self.prefix + str(obj), *args, **kwargs)
		return wrappedFunc
