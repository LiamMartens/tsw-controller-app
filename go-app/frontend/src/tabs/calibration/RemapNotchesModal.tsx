import { main } from "../../../wailsjs/go/models";
import { CalibrationModalForm } from "./CalibrationModalForm";
import { Modal, ModalContentProps } from "../../components";
import { RemapNotchesModalForm } from "./RemapNotchesModalForm";

type Props = {
  controller: main.Interop_GenericController | null;
  onClose: () => void;
};

const RemapNotchesModalContent = ({
  openState: controller,
  onClose,
}: ModalContentProps<main.Interop_GenericController>) => {
  return (
    <RemapNotchesModalForm
      controller={controller}
      onClose={onClose}
    />
  );
};

export const RemapNotchesModal = ({ controller, onClose }: Props) => {
  return (
    <Modal
      openState={controller ?? false}
      onClose={onClose}
      Component={RemapNotchesModalContent}
    />
  );
};
