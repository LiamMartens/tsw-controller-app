import { createPortal } from "react-dom";
import z from "zod";
import { CONDITION_OPERATORS } from "../utils";
import { useId, useMemo } from "react";
import { t } from "../../utils";

type Operator = (typeof CONDITION_OPERATORS)[number]["value"];

const conditionSchema = z.object({
  control: z.string(),
  operator: z.enum(
    CONDITION_OPERATORS.map((c) => c.value) as [Operator, ...Operator[]],
  ),
  value: z.number(),
});

type Props = {
  value: z.infer<typeof conditionSchema>[];
  onChange: (value: z.infer<typeof conditionSchema>[]) => void;
};

type ModalProps = Props & {
  dialogId: string;
};

const EditConditionsModalComponent = ({ dialogId }: ModalProps) => {
  return (
    <dialog id={dialogId} className="modal">
      <div className="modal-box">
        <h3 className="font-bold text-lg">{t("Edit conditions")}</h3>
        <p className="py-4"></p>
        <form method="dialog" className="modal-action">
          <button className="btn btn-sm btn-primary">{t("Save")}</button>
        </form>
      </div>
    </dialog>
  );
};

export const EditConditionsModal = ({ value, onChange }: Props) => {
  const id = useId();
  const dialogId = useMemo(
    () => `edit-conditions-dialog-` + id.replace(/[^\w]+/g, "-"),
    [id],
  );

  return {
    render: () =>
      createPortal(
        <EditConditionsModalComponent
          dialogId={dialogId}
          value={value}
          onChange={onChange}
        />,
        document.body,
      ),
  };
};
