import { apiRequest } from "@/common/api";
import { type Config } from "@/config/types";

export async function getConfig() {
	return await apiRequest<Config>("GET", "config");
}

export async function putConfig(config: Config) {
	return await apiRequest<Config>("PUT", "config", config);
}

export async function getConfigAutostart() {
	return await apiRequest<boolean>("GET", "config/autostart");
}

export async function putConfigAutostart(enabled: boolean) {
	return await apiRequest<boolean>("PUT", "config/autostart", enabled);
}
