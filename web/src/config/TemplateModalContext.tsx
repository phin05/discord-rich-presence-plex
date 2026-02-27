import { createContext, useContext } from "react";

export type OpenTemplateModal = (fieldName: string) => void;
export const TemplateModalContext = createContext<OpenTemplateModal | null>(null);
export function useTemplateModal() {
	const openTemplateModal = useContext(TemplateModalContext);
	if (!openTemplateModal) {
		throw new Error("useTemplateModal must be used within a TemplateModalProvider");
	}
	return openTemplateModal;
}
