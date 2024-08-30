from .config import config
from PIL import Image
from typing import Optional
from utils.logging import logger
import io
import models.imgur
import requests

def uploadToImgur(url: str) -> Optional[str]:
	try:
		originalImageBytesIO = io.BytesIO(requests.get(url).content)
		originalImage = Image.open(originalImageBytesIO)
		newImage = Image.new("RGB", originalImage.size)
		newImage.putdata(originalImage.getdata()) # pyright: ignore[reportArgumentType]
		maxSize = config["display"]["posters"]["maxSize"]
		if maxSize:
			newImage.thumbnail((maxSize, maxSize))
		newImageBytesIO = io.BytesIO()
		newImage.save(newImageBytesIO, subsampling = 0, quality = 90, format = "JPEG")
		data: models.imgur.UploadResponse = requests.post(
			"https://api.imgur.com/3/image",
			headers = { "Authorization": f"Client-ID {config['display']['posters']['imgurClientID']}" },
			files = { "image": newImageBytesIO.getvalue() }
		).json()
		if not data["success"]:
			raise Exception(data["data"]["error"])
		return data["data"]["link"]
	except:
		logger.exception("An unexpected error occured while uploading an image to Imgur")
