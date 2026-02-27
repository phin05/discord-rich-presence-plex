import { getConfigAutostart, putConfigAutostart } from "@/config/api";
import { Divider, Switch } from "@mantine/core";
import { notifications } from "@mantine/notifications";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

export function AutostartSwitch() {
	const queryClient = useQueryClient();

	const autostart = useQuery({
		queryKey: ["autostart"],
		queryFn: getConfigAutostart,
	});

	const setAutostart = useMutation({
		mutationKey: ["setAutostart"],
		mutationFn: putConfigAutostart,
		meta: { errorTitle: "Failed to update autostart status" },
		onSuccess: (newData) => {
			queryClient.setQueryData(["autostart"], newData);
			notifications.show({
				color: "green",
				title: `Autostart ${newData ? "enabled" : "disabled"}`,
				message: "",
			});
		},
	});

	return (
		autostart.data !== undefined && (
			<>
				<Divider orientation="vertical" />
				<Switch
					checked={autostart.data}
					disabled={setAutostart.isPending}
					label="Launch app on system startup"
					onChange={(event) => {
						setAutostart.mutate(event.currentTarget.checked);
					}}
				/>
			</>
		)
	);
}
