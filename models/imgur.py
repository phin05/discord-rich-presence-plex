from typing import TypedDict

class ImgurResponse(TypedDict):
	success: bool
	status: int

class ImgurUploadResponseData(TypedDict):
	error: str
	link: str

class ImgurUploadResponse(ImgurResponse):
	data: ImgurUploadResponseData
