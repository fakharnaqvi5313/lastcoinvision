import { accent, navy } from "@lastcoinvision/ui-tokens";
import { StyleSheet } from "react-native";

export { accent, navy };

export const styles = StyleSheet.create({
  screen: {
    flex: 1,
    backgroundColor: navy[950],
    padding: 24,
    justifyContent: "center",
  },
  card: {
    backgroundColor: navy[900],
    borderColor: navy[700],
    borderWidth: 1,
    borderRadius: 12,
    padding: 20,
  },
  title: {
    fontSize: 22,
    fontWeight: "700",
    color: accent[500],
    marginBottom: 8,
  },
  body: {
    fontSize: 14,
    color: navy[200],
    marginBottom: 16,
    lineHeight: 20,
  },
  label: {
    fontSize: 13,
    color: navy[200],
    marginBottom: 6,
  },
  input: {
    backgroundColor: navy[950],
    borderColor: navy[600],
    borderWidth: 1,
    borderRadius: 6,
    color: navy[50],
    padding: 12,
    fontSize: 15,
    marginBottom: 14,
  },
  textArea: {
    minHeight: 90,
    textAlignVertical: "top",
  },
  buttonRow: {
    flexDirection: "row",
    gap: 12,
    marginTop: 8,
  },
  primaryButton: {
    backgroundColor: accent[500],
    borderRadius: 6,
    paddingVertical: 12,
    paddingHorizontal: 18,
    alignItems: "center",
    flex: 1,
  },
  primaryButtonDisabled: {
    opacity: 0.5,
  },
  primaryButtonText: {
    color: navy[950],
    fontWeight: "700",
    fontSize: 15,
  },
  secondaryButton: {
    borderColor: navy[600],
    borderWidth: 1,
    borderRadius: 6,
    paddingVertical: 12,
    paddingHorizontal: 18,
    alignItems: "center",
    flex: 1,
  },
  secondaryButtonText: {
    color: navy[50],
    fontWeight: "600",
    fontSize: 15,
  },
  mnemonicBox: {
    backgroundColor: navy[950],
    borderColor: accent[600],
    borderWidth: 1,
    borderStyle: "dashed",
    borderRadius: 8,
    padding: 16,
    marginBottom: 16,
  },
  mnemonicText: {
    color: accent[300],
    fontSize: 16,
    lineHeight: 28,
    fontFamily: "monospace",
  },
  warning: {
    color: accent[400],
    fontSize: 13,
    marginBottom: 12,
  },
  error: {
    color: "#ff9c9c",
    fontSize: 13,
    marginBottom: 8,
  },
  addressRow: {
    backgroundColor: navy[950],
    borderRadius: 8,
    padding: 12,
    marginBottom: 12,
  },
  addressLabel: {
    fontSize: 11,
    textTransform: "uppercase",
    letterSpacing: 0.5,
    color: navy[300],
    marginBottom: 4,
  },
  addressValue: {
    fontFamily: "monospace",
    fontSize: 13,
    color: navy[50],
  },
  muted: {
    color: navy[300],
    fontSize: 13,
  },
  linkButton: {
    marginTop: 4,
  },
  linkButtonText: {
    color: accent[500],
    fontSize: 13,
    textDecorationLine: "underline",
  },
  pubkeyBox: {
    backgroundColor: navy[900],
    borderRadius: 6,
    padding: 8,
    marginTop: 8,
    marginBottom: 4,
  },
  pubkeyText: {
    fontFamily: "monospace",
    fontSize: 11,
    color: navy[50],
  },
  checkboxRow: {
    flexDirection: "row",
    alignItems: "center",
    gap: 10,
    marginBottom: 8,
  },
  tabs: {
    flexDirection: "row",
    gap: 8,
    marginVertical: 12,
  },
  tab: {
    backgroundColor: navy[800],
    borderRadius: 6,
    paddingVertical: 8,
    paddingHorizontal: 14,
  },
  tabActive: {
    backgroundColor: accent[600],
  },
  tabText: {
    color: navy[200],
    fontSize: 13,
    fontWeight: "600",
  },
  tabTextActive: {
    color: navy[950],
  },
  txResult: {
    backgroundColor: navy[950],
    borderRadius: 8,
    padding: 10,
    marginTop: 12,
  },
  txResultText: {
    fontFamily: "monospace",
    fontSize: 11,
    color: navy[100],
  },
});
