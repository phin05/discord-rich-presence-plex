from typing import TypedDict

class Response(TypedDict):
	success: bool
	status: int

class UploadResponseData(TypedDict):
	error: str
	link: str

class UploadResponse(Response):
	data: UploadResponseData
