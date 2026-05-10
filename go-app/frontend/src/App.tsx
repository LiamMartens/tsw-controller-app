import { useForm } from "react-hook-form";
import { MainTab } from "./tabs/main";
import { CalibrationTab } from "./tabs/calibration";
import { LogsTab } from "./tabs/logs";
import { CabDebuggerTab } from "./tabs/cabdebugger";
import { SelfUpdateBanner } from "./SelfUpdateBanner";
import { ExploreTab } from "./tabs/explore";
import { SettingsTab } from "./tabs/settings";
import { Suspense } from "react";
import { ErrorBoundary } from "react-error-boundary";

const App = () => {
  const tabsForm = useForm<{
    tab:
      | "main"
      | "explore"
      | "calibration"
      | "cab_debugger"
      | "logs"
      | "settings";
  }>({
    defaultValues: { tab: "main" },
  });
  const tab = tabsForm.watch("tab");

  const handleOpenCabDebuggerTab = () => {
    tabsForm.setValue("tab", "cab_debugger");
  };

  return (
    <div className="p-2">
      <SelfUpdateBanner />

      <div className="sticky top-2 tabs tabs-box z-10">
        <input
          type="radio"
          className="tab"
          aria-label="Main"
          value="main"
          {...tabsForm.register("tab", { value: "main" })}
        />
        <input
          type="radio"
          className="tab"
          aria-label="Explore"
          value="explore"
          {...tabsForm.register("tab", { value: "explore" })}
        />
        <input
          type="radio"
          className="tab"
          aria-label="Cab Debugger"
          value="cab_debugger"
          {...tabsForm.register("tab", { value: "cab_debugger" })}
        />
        <input
          type="radio"
          className="tab"
          aria-label="Calibration"
          value="calibration"
          {...tabsForm.register("tab", { value: "calibration" })}
        />
        <input
          type="radio"
          className="tab"
          aria-label="Logs"
          value="logs"
          {...tabsForm.register("tab", { value: "logs" })}
        />
        <input
          type="radio"
          className="tab"
          aria-label="Settings"
          value="settings"
          {...tabsForm.register("tab", { value: "settings" })}
        />
      </div>

      <div className="p-2">
        <ErrorBoundary
          fallback={
            <p className="text-error text-center py-20">An error occured</p>
          }
        >
          <Suspense>
            {tab === "main" && (
              <MainTab onOpenCabDebuggerTab={handleOpenCabDebuggerTab} />
            )}
            {tab === "explore" && <ExploreTab />}
            {tab === "cab_debugger" && <CabDebuggerTab />}
            {tab === "calibration" && <CalibrationTab />}
            {tab === "logs" && <LogsTab />}
            {tab === "settings" && <SettingsTab />}
          </Suspense>
        </ErrorBoundary>
      </div>
    </div>
  );
};

export default App;
