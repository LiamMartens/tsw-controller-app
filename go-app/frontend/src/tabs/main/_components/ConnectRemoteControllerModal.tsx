import { MutableRefObject, Suspense, useRef } from "react";
import QRCode from "react-qr-code";
import { GetDeviceIP } from "../../../../wailsjs/go/main/App";
import { BrowserOpenURL } from "../../../../wailsjs/runtime/runtime";
import { useDeviceIP } from "../../../swr";
import { ErrorBoundary } from "react-error-boundary";
import { Modal } from "../../../components";

type Props = {
  open: boolean;
  onClose: () => void;
};

const ConnectRemoteControllerModalContent = () => {
  const { data: deviceIP } = useDeviceIP();

  const handleOpenAppLink = () => {
    BrowserOpenURL(
      "https://github.com/LiamMartens/tsw-controller-app/releases",
    );
  };

  const handleOpenGuideLink = () => {
    BrowserOpenURL(
      "https://tsw-controller-app.vercel.app/docs/setting-up-virtual-controller",
    );
  };

  return (
    <div>
      <form method="dialog">
        <button className="btn btn-sm btn-circle btn-ghost absolute right-2 top-2">
          ✕
        </button>
      </form>

      <div>
        <div className="flex justify-center">
          {deviceIP && (
            <div className="p-4 bg-white rounded-md">
              <QRCode
                value={JSON.stringify({
                  connection: {
                    ip: deviceIP,
                    port: 63241,
                  },
                })}
              />
            </div>
          )}
          {!deviceIP && (
            <div className="alert alert-error">
              Could not determine connection address for remote controller
            </div>
          )}
        </div>
        <div className="alert mt-2">
          <span>
            Use the{" "}
            <button className="link" onClick={handleOpenAppLink}>
              TSW Virtual Controller app
            </button>{" "}
            to scan the QR code and connect your android device.
          </span>
        </div>
        <div className="alert mt-2">
          <span>
            Don't know how to set-up a virtual controller? Follow the{" "}
            <button className="link" onClick={handleOpenGuideLink}>
              online guide
            </button>
          </span>
        </div>
      </div>
    </div>
  );
};

export const ConnectRemoteControllerModal = ({ open, onClose }: Props) => {
  return (
    <Modal
      openState={open}
      onClose={onClose}
      Component={ConnectRemoteControllerModalContent}
    />
  );
};
