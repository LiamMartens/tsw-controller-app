import { MutableRefObject, useRef } from "react";

type Props = {
  dialogRef: MutableRefObject<HTMLDialogElement | null>;
};

export const EasyMapperModal = ({ dialogRef }: Props) => {
  const ref = useRef<HTMLDialogElement | null>(null);

  const handleRef = (d: HTMLDialogElement | null) => {
    ref.current = d;
    dialogRef.current = d;
  };

  const handleClose = () => {
    ref.current?.close();
  };

  return (
    <dialog ref={handleRef} className="modal modal-s">
      <div className="modal-box w-11/12 max-w-5xl">
      </div>
    </dialog>
  );
};
