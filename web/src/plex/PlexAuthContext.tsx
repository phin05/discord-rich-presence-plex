import { createContext, useContext } from "react";

export type PlexAuthOnAdd = (token: string, serverName: string) => void;
export type PlexAuthOnCancel = () => void;
export type PlexAuth = (onAdd: PlexAuthOnAdd, onCancel?: PlexAuthOnCancel) => void;
export const PlexAuthContext = createContext<PlexAuth | null>(null);
export function usePlexAuth() {
	const plexAuth = useContext(PlexAuthContext);
	if (!plexAuth) {
		throw new Error("usePlexAuth must be used within a PlexAuthProvider");
	}
	return plexAuth;
}
