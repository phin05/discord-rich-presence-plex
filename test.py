from services import DiscordRpcService
import time

discordRpcService = DiscordRpcService()
discordRpcService.connect()
discordRpcService.setActivity({
	"details": "details",
	"state": "state",
	"assets": {
		"large_text": "large_text",
		"large_image": "logo",
		"small_text": "small_text",
		"small_image": "playing",
	},
})
time.sleep(60)
