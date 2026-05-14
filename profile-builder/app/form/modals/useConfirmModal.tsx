import { ReactNode, SubmitEvent, useCallback, useId, useMemo } from "react";
import { t } from "../../utils";
import { createPortal } from "react-dom";

type ModalProps = {
  title: string;
  body: ReactNode;
  onConfirm: () => void;
};

type ConfirmDialogProps = {
  dialogId: string;
  title: string;
  body: ReactNode;
  onSubmit: (event: SubmitEvent<HTMLDialogElement>) => void;
};

const ConfirmDialog = ({
  dialogId,
  title,
  body,
  onSubmit,
}: ConfirmDialogProps) => {
  return (
    <dialog id={dialogId} className="modal" onSubmit={onSubmit}>
      <div className="modal-box">
        <h3 className="font-bold text-lg">{title}</h3>
        <p className="py-4">{body}</p>
        <form method="dialog" className="modal-action">
          <button className="btn btn-sm" name="cancel" value="cancel">
            {t("Cancel")}
          </button>
          <button
            className="btn btn-sm btn-primary"
            name="action"
            value="confirm"
          >
            {t("Confirm")}
          </button>
        </form>
      </div>
    </dialog>
  );
};

export const useConfirmModal = ({ title, body, onConfirm }: ModalProps) => {
  const id = useId();
  const dialogId = useMemo(() => `dialog-` + id.replace(/[^\w]+/g, "-"), [id]);

  const handleSubmit = useCallback(
    (event: SubmitEvent<HTMLDialogElement>) => {
      event.stopPropagation();
      const submitter = event.nativeEvent.submitter as HTMLButtonElement | null;
      if (submitter?.value === "cancel") return;
      onConfirm();
    },
    [onConfirm],
  );

  const confirm = useCallback(() => {
    (document.getElementById(dialogId) as HTMLDialogElement).showModal();
  }, [dialogId]);

  return {
    confirm,
    render: () =>
      createPortal(
        <ConfirmDialog
          dialogId={dialogId}
          title={title}
          body={body}
          onSubmit={handleSubmit}
        />,
        document.body,
      ),
  };
};
