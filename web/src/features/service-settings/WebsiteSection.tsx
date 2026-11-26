// SPDX-License-Identifier: AGPL-3.0-only

// LuxRail: Settings > Website. Marks a service as a web frontend and keeps the
// https address people open it at, shown as a link on the service card. Both
// save at once (PATCH .../info); they never stage a change or redeploy.

import type { Service } from "../../api/types.gen";
import { patch } from "../service/client";
import { toastError } from "../../kit";
import { useSettings } from "./context";
import { SaveInput, SettingField, ToggleRow } from "./ui";
import { websiteError } from "./website";

export default function WebsiteSection() {
  const { base, svc, onService } = useSettings();
  const cfg = svc.config;
  const save = async (b: { webFrontend?: boolean; website?: string }) => {
    try { onService(await patch<Service>(`${base}/info`, b)); } catch (e) { toastError(e); throw e; }
  };
  return (
    <>
      <SettingField title="Web Frontend" keywords="website frontend link url open"
        description="Mark this service as a website to show a link to it on its card. The address can be one this install does not serve itself, such as a Cloudflare tunnel.">
        <ToggleRow label="This is a web frontend" checked={!!cfg.webFrontend} onChange={(v) => void save({ webFrontend: v }).catch(() => undefined)} />
      </SettingField>
      {cfg.webFrontend && (
        <SettingField title="Website" keywords="website url link https address">
          <SaveInput label="Website" value={cfg.website || ""} placeholder="https://app.example.com" mono
            validate={websiteError} onSave={(v) => save({ website: v })} />
        </SettingField>
      )}
    </>
  );
}
