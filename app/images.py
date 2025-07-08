from app import cache, logger, config
from PIL import Image, ImageOps
from typing import Optional
from typing import TypedDict
import io
import requests
import time

def upload(key: str, url: str) -> Optional[str]:
	cachedValue = cache.get(key)
	if cachedValue:
		return cachedValue
	originalImageBytesIO = io.BytesIO(requests.get(url).content)
	originalImage = Image.open(originalImageBytesIO).convert("RGBA")
	newImage = Image.new("RGBA", originalImage.size)
	newImage.putdata(originalImage.getdata()) # pyright: ignore[reportUnknownArgumentType,reportUnknownMemberType]
	if newImage.width != newImage.height and config.config["display"]["posters"]["fit"]:
		longestSideLength = max(newImage.width, newImage.height)
		newImage = ImageOps.pad(newImage, (longestSideLength, longestSideLength), color = (0, 0, 0, 0))
	maxSize = config.config["display"]["posters"]["maxSize"]
	if maxSize:
		newImage.thumbnail((maxSize, maxSize))
	newImageBytesIO = io.BytesIO()
	newImage.save(newImageBytesIO, subsampling = 0, quality = 90, format = "PNG")
	pngBytes = newImageBytesIO.getvalue()
	try:
		if config.config["display"]["posters"]["imgurClientID"]:
			uploadedImageUrl = uploadToImgur(pngBytes)
			cache.set(key, uploadedImageUrl, 0)
		else:
			uploadedImageUrl = uploadToLitterbox(pngBytes)
			cache.set(key, uploadedImageUrl, int(time.time()) + (72 * 60 * 60))
		return uploadedImageUrl
	except:
		logger.exception("An unexpected error occured while uploading an image")

def uploadToLitterbox(pngBytes: bytes) -> str:
	logger.debug("Uploading image to Litterbox")
	response = requests.post(
		"https://litterbox.catbox.moe/resources/internals/api.php",
		data = { "reqtype": "fileupload", "time": "72h" },
		files = { "fileToUpload": ("image.png", pngBytes) }
	)
	logger.debug("HTTP %d, %s, %s", response.status_code, response.headers, response.text.strip())
	return response.text.strip()

class ImgurUploadResponseData(TypedDict):
	error: str
	link: str

class ImgurUploadResponse(TypedDict):
	success: bool
	status: int
	data: ImgurUploadResponseData

def uploadToImgur(pngBytes: bytes) -> str:
	logger.debug("Uploading image to Imgur")
	response = requests.post(
		"https://api.imgur.com/3/image",
		headers = { "Authorization": f"Client-ID {config.config['display']['posters']['imgurClientID']}" },
		files = { "image": pngBytes }
	)
	logger.debug("HTTP %d, %s, %s", response.status_code, response.headers, response.text.strip())
	data: ImgurUploadResponse = response.json()
	if not data["success"]:
		raise Exception(data["data"]["error"])
	return data["data"]["link"]
