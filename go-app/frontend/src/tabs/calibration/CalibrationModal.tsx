import { main } from "../../../wailsjs/go/models";
import { CalibrationModalForm } from "./CalibrationModalForm";
import { Modal, ModalContentProps } from "../../components";

type Props = {
  controller: main.Interop_GenericController | null;
  onClose: () => void;
};

const CalibrationModalContent = ({
  openState: controller,
  onClose,
}: ModalContentProps<main.Interop_GenericController>) => {
  return (
    <CalibrationModalForm
      key={controller.UniqueID}
      controller={controller}
      onClose={onClose}
    />
  );
};

export const CalibrationModal = ({ controller, onClose }: Props) => {
  return (
    <Modal
      openState={controller ?? false}
      onClose={onClose}
      Component={CalibrationModalContent}
    />
  );
};
