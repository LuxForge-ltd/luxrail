// SPDX-License-Identifier: AGPL-3.0-only

// LuxRail: the user menu's Install CLI window. The server hands out the CLI
// build of its own version; the CLI then keeps itself at the server's version.
import { useEffect, useMemo, useState } from "react";
import { Download } from "lucide-react";
import { api } from "../../api/client";
import { Modal, Select } from "../../kit";
import { CopyField, Loading, Notice, useLoad } from "../settings/ui";
import { detectDownload, installSteps, platformLabel } from "./platform";

export function InstallCLI({ onClose }: { onClose: () => void }) {
  const info = useLoad(api.cli, []);
  const [pick, setPick] = useState("");
  const downloads = useMemo(() => info.data?.downloads ?? [], [info.data]);
  useEffect(() => {
    if (!pick && downloads.length) {
      const nav = navigator as Navigator & { userAgentData?: { platform?: string } };
      const d = detectDownload(downloads, navigator.userAgent, nav.userAgentData?.platform ?? navigator.platform ?? "");
      if (d) setPick(`${d.os}/${d.arch}`);
    }
  }, [downloads, pick]);
  const d = downloads.find((x) => `${x.os}/${x.arch}` === pick);
  const login = `luxrail login ${window.location.origin}`;
  return (
    <Modal title="Install the CLI" onClose={onClose} width={560}
      description="Deploy and manage apps from your terminal with luxrail.">
      {!info.data ? <Loading error={info.error} /> : !info.data.available ? (
        <Notice>This server is a development build, so it has no CLI to download. Install the CLI from a LuxRail release, or build it from source.</Notice>
      ) : (
        <div className="space-y-5 pb-2 text-body">
          <section className="space-y-2">
            <h3 className="font-medium text-foreground">1. Download luxrail {info.data.version}</h3>
            <div className="flex flex-wrap items-center gap-2">
              <Select aria-label="System" value={pick} onChange={(e) => setPick(e.target.value)}
                options={downloads.map((x) => ({ value: `${x.os}/${x.arch}`, label: platformLabel(x) }))} />
              {d && (
                <a href={d.path} download={d.file}
                  className="inline-flex h-[34px] items-center gap-2 rounded-md border border-accent bg-accent px-3 text-sm text-accent-fg transition-colors hover:border-accent-hover hover:bg-accent-hover">
                  <Download size={16} />Download
                </a>
              )}
            </div>
            {d && <p className="text-caption text-fg-3">{d.file}</p>}
          </section>
          {d && (
            <section className="space-y-2">
              <h3 className="font-medium text-foreground">2. Put it on your PATH</h3>
              {installSteps(d.os, d.file).map((s) => <CopyField key={s} value={s} />)}
              {d.os === "windows"
                ? <p className="text-caption text-fg-3">Open a new terminal afterwards so it sees the new PATH.</p>
                : <p className="text-caption text-fg-3">Add ~/.local/bin to your PATH if your shell does not already have it.</p>}
            </section>
          )}
          <section className="space-y-2">
            <h3 className="font-medium text-foreground">3. Log in to this console</h3>
            <CopyField value={login} />
            <p className="text-caption text-fg-3">
              The CLI updates itself to this server's version when it runs. Set LUXRAIL_NO_UPDATE=1 to turn that off, or run luxrail self-update to update now.
            </p>
          </section>
        </div>
      )}
    </Modal>
  );
}
