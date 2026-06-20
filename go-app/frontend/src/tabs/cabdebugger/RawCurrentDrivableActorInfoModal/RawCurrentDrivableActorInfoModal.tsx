import { useForm } from "react-hook-form";
import { main } from "../../../../wailsjs/go/models";
import { Modal, ModalContentProps } from "../../../components";

type Props = {
  open: boolean;
  onClose: () => void;
};

type ContentProps = ModalContentProps<boolean>;

const RawCurrentDrivableActorInfoModalContent = ({
  openState: controlName,
  onClose,
}: ContentProps) => {
  return <></>;
};

export const RawCurrentDrivableActorInfoModal = ({ open, onClose }: Props) => {
  return (
    <Modal
      openState={open}
      onClose={onClose}
      Component={RawCurrentDrivableActorInfoModalContent}
    />
  );
};
