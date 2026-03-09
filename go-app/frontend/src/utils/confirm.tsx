import { StrictMode, FormEvent, SyntheticEvent } from "react";
import { createRoot } from "react-dom/client";

type ConfirmInput = {
  id: string;
  title: string;
  message: string;
  actions: [string, string];
  onCancel?: () => void;
  onConfirm: () => void;
};

export const confirm = (input: ConfirmInput) => {
  const container = document.createElement("div");
  document.body.appendChild(container);

  const handleClose = (event: SyntheticEvent<HTMLDialogElement>) => {
    event.currentTarget.addEventListener("transitionend", container.remove);
    if (event.currentTarget.returnValue === input.actions[0]) {
      input.onCancel?.();
    } else if (event.currentTarget.returnValue === input.actions[1]) {
      input.onConfirm();
    }
  };

  const ConfirmDialog = () => {
    return (
      <dialog
        id={input.id}
        className="modal"
        ref={(ref) => ref?.showModal()}
        onClose={handleClose}
      >
        <div className="modal-box">
          <h3 className="text-lg font-bold">{input.title}</h3>
          <p className="py-4">{input.message}</p>
          <div className="modal-action">
            <form method="dialog" className="flex gap-2 justify-end">
              {input.actions.map((action) => (
                <button
                  key={action}
                  name="action"
                  value={action}
                  className="btn btn-sm"
                >
                  {action}
                </button>
              ))}
            </form>
          </div>
        </div>
      </dialog>
    );
  };

  const root = createRoot(container);
  root.render(
    <StrictMode>
      <ConfirmDialog />
    </StrictMode>,
  );
};
