import { MutableRefObject, useRef } from "react";
import { Modal } from "../../components";

type Props = {
  open: boolean;
  onClose: () => void;
};

const EasyMapperModalContent = () => {
  return null;
};

export const EasyMapperModal = ({ open, onClose }: Props) => {
  return (
    <Modal
      openState={open}
      onClose={onClose}
      Component={EasyMapperModalContent}
    />
  );
};
