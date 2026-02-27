import { App } from "@/App";
import { type ApiError } from "@/common/api";
import { PlexAuthProgressModal } from "@/plex/PlexAuthProvider";
import { MantineProvider } from "@mantine/core";
import "@mantine/core/styles.css";
import { notifications, Notifications } from "@mantine/notifications";
import "@mantine/notifications/styles.css";
import { MutationCache, QueryCache, QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { StrictMode } from "react";
import { createRoot } from "react-dom/client";

declare module "@tanstack/react-query" {
	interface Register {
		defaultError: ApiError;
	}
}

function showErrorNotification(error: ApiError, meta?: Record<string, unknown>) {
	if (typeof meta?.errorTitle === "string") {
		notifications.show({
			color: "red",
			title: meta.errorTitle,
			message: `${error.message}${error.details ? ` (additional details: ${error.details.join(", ")})` : ""}`,
		});
	}
}

const queryClient = new QueryClient({
	defaultOptions: {
		queries: {
			retry: false,
			refetchOnWindowFocus: false,
			refetchOnReconnect: false,
			refetchOnMount: true,
		},
		mutations: {
			retry: false,
		},
	},
	queryCache: new QueryCache({
		onError: (error, query) => {
			showErrorNotification(error, query.meta);
		},
	}),
	mutationCache: new MutationCache({
		onError: (error, _variables, _onMutateResult, mutation, _context) => {
			showErrorNotification(error, mutation.meta);
		},
	}),
});

const params = new URLSearchParams(window.location.search);
if (params.has("plexAuthCallback") && window.opener) {
	window.close();
} else {
	// eslint-disable-next-line @typescript-eslint/no-non-null-assertion
	createRoot(document.getElementById("root")!).render(
		<StrictMode>
			<QueryClientProvider client={queryClient}>
				<MantineProvider defaultColorScheme="dark">
					<Notifications autoClose={10000} position="top-center" transitionDuration={500} />
					{params.has("plexAuthWaiting") ? <PlexAuthProgressModal statusText="Redirecting to Plex..." /> : <App />}
				</MantineProvider>
			</QueryClientProvider>
		</StrictMode>,
	);
}
