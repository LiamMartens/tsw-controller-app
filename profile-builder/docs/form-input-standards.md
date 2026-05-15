# Form & Input Implementation Standards

Derived from the form components in `app/form/`.

---

## 1. Form Engine

- **React Hook Form** is the sole form library. All forms use `useForm` with `zodResolver`.
- **Zod** schemas drive validation. Each field-level sub-form (e.g. `MomentaryFields`) defines its own Zod schema and resolver.
- Top-level forms (`ProfileForm`) use a single `profileSchema` for the entire profile.
- `mode: "onChange"` is used for sub-forms to get real-time validation feedback.

```tsx
const form = useForm<z.infer<typeof schema>>({
  mode: "onChange",
  resolver: zodResolver(schema),
  defaultValues: safeValue,
});
```

---

## 2. Field Wrappers

### `BaseField`

A `<fieldset>` wrapper used for **every** input. Always use it — never render a raw `<input>` or `<select>` without it.

| Prop        | Required | Purpose                                                       |
| ----------- | -------- | ------------------------------------------------------------- |
| `legend`    | No       | Fieldset legend text                                          |
| `label`     | No       | Descriptive helper text (rendered as `<p className="label">`) |
| `error`     | No       | Validation error message (rendered in `text-error`)           |
| `className` | No       | Extra classes on the `<fieldset>`                             |
| `children`  | Yes      | The actual input element                                      |

**Styling conventions:**

- Inputs use `input input-bordered w-full` (DaisyUI).
- Error state adds `input-error`.
- Labels use `label whitespace-normal` for wrapping.
- Error paragraphs use `label text-error whitespace-normal`.

```tsx
<BaseField legend={t("Label")} label={t("Helper text")} error={error}>
  <input
    className={clsx("input input-bordered w-full", error && "input-error")}
    {...form.register("fieldName")}
  />
</BaseField>
```

### `FieldGroup`

A styled `<fieldset>` variant for **grouping related inputs** (e.g. action selectors). Adds a background, border, padding, and rounded corners.

| Prop       | Required | Purpose                 |
| ---------- | -------- | ----------------------- |
| `legend`   | No       | Group legend            |
| `label`    | No       | Group-level helper text |
| `children` | Yes      | Input elements inside   |

---

## 3. Input Types & Patterns

### Text / Number Inputs

- Use `form.register("path", { valueAsNumber: true })` for numeric fields.
- Always include `placeholder` for guidance.
- Wire validation errors via `form.formState.errors.path?.message`.

### Select Inputs

- Use `form.register("path")` on `<select>`.
- Apply `select-error` class when there is an error.

```tsx
<select className={clsx("select w-full", error && "select-error")} {...form.register("match")}>
  <option value="exceeds">{t("Exceeds")}</option>
  <option value="equals">{t("Exact match")}</option>
</select>
```

### Checkbox Inputs

- Use `form.register("path")` on `<input type="checkbox">`.
- For conditional rendering based on checkbox state, use `form.setValue` with `shouldDirty: true, shouldTouch: true`.

### JSON Textarea (`JsonTextareaInput`)

- A dedicated component for editing JSON objects inline.
- Accepts `value` (object or null) and `onChange` (callback receiving parsed object).
- Internally manages a `textarea` with `textarea textarea-bordered h-20 font-mono text-sm resize-y`.
- Parses on every keystroke (try/catch to avoid crashes on invalid JSON).

---

## 4. Dynamic / Array Fields

- Use `useFieldArray` from React Hook Form for repeating sections.
- Destructure `{ fields, append, remove }` from `useFieldArray`.
- **Always use `field.id`** as the React key — never use the array index.
- `append` receives a default object shape. `remove` takes the array index.

```tsx
const { fields, append, remove } = useFieldArray({
  control: form.control,
  name: "controls",
});

{
  fields.map((field, index) => (
    <ControlCard key={field.id} form={form} index={index} onRemove={() => remove(index)} />
  ));
}
```

---

## 5. Nested / Controller Fields

- Use `<Controller>` from React Hook Form for deeply nested or complex fields.
- `Controller` provides `field` (value, onChange, onBlur, name) and `fieldState` (error, invalid, isDirty, etc.).
- For sub-forms inside a `Controller`, lift the value out via `field.value` and push changes via `field.onChange`.

```tsx
<Controller
  control={form.control}
  name={`controls.${index}.assignments.${assignmentIndex}`}
  render={({ field }) => <MomentaryFieldsContent value={field.value} onChange={field.onChange} />}
/>
```

### Sub-forms inside Controller

When a `Controller` renders a component that owns its own `useForm`:

1. Validate the incoming `value` with a sub-schema via `useMemo` for memoization.
2. Provide `defaultValues` from the validated data, with sensible fallbacks.
3. Sync the sub-form's output back to the parent via `form.watch(() => form.handleSubmit(onChange)())`.
4. Use `useRef` + `useEffect` to avoid stale closure issues with the `onChange` callback.

```tsx
const onChangeRef = useRef(onChange);
onChangeRef.current = onChange;
useEffect(() => {
  form.watch(() => form.handleSubmit(onChange)());
}, [form, onChangeRef]);
```

---

## 6. Collapsible Sections

- Use the DaisyUI `collapse` pattern: `<div className="collapse collapse-arrow">` with a hidden `<input type="checkbox" className="peer" />` and `<div className="collapse-title">`.
- Sections are used for grouping related but non-critical information (e.g. controller info, rail classes).
- Collapse content uses `bg-base-100 border-base-300 border`.

---

## 7. Action Selector Pattern

- Complex action types are abstracted behind an `ActionSelector` component.
- Each action type has its own fields component (e.g. `KeysActionFields`, `ApiControlActionFields`).
- Actions are stored as nested objects (e.g. `action_activate`, `action_deactivate`).
- Use `form.setValue` with `{ shouldDirty: true, shouldTouch: true }` when updating action values.

---

## 8. Confirmation Modals

- Use `useConfirmModal` hook for destructive actions (e.g. removing a control).
- The hook returns `{ confirm, render }`. Call `confirm()` to trigger, and `render` as a component to render the modal.

```tsx
const { confirm, render: ConfirmDeleteComponent } = useConfirmModal({
  title: t("Are you sure?"),
  body: t("Are you sure you want to remove this control?"),
  onConfirm: () => onRemove(),
});
```

---

## 9. i18n / Translation

- All user-facing strings go through the `t()` utility.
- `t()` is imported from `../utils` (or `../../../utils` for deeper nesting).
- No hardcoded strings in JSX — always wrap in `t()`.

---

## 10. Layout & Card Structure

- Top-level form sections live inside `<div className="card bg-base-100 border border-base-300 shadow-xl">`.
- Card bodies use `<div className="card-body">` with `<h2 className="card-title">` for section headings.
- Vertical spacing between major sections: `space-y-6`.
- Spacing between inputs within a section: `space-y-4` or `gap-2` / `gap-4` as appropriate.

---

## 11. Submit Handling

- Top-level forms use `form.handleSubmit(handleSubmit)` on the `<form>` element.
- Sub-forms (inside `Controller`) push changes via `onChange` rather than explicit submit.
- Save button uses `btn btn-primary`.

---

## 12. Error Display

- Errors are always displayed **below** the input, inside a `<p className="label text-error whitespace-normal">`.
- For array fields, access errors via `form.formState.errors.arrayName?.[index]?.field?.message`.
- Input elements get an error class (e.g. `input-error`, `select-error`) when `form.formState.errors.path` exists.

---

## 13. Component File Organization

```
app/form/
├── ProfileForm.tsx          # Top-level form, schema, submit
├── ProfileHeader.tsx        # Profile metadata fields
├── ControlsList.tsx         # Dynamic control array
├── ControlCard.tsx          # Single control card (collapsible)
├── inputs/
│   ├── index.ts             # Barrel export
│   ├── BaseField.tsx        # Universal field wrapper
│   ├── FieldGroup.tsx       # Grouped field wrapper
│   └── JsonTextareaInput.tsx # JSON editing textarea
├── actions/
│   ├── ActionSelector.tsx   # Action type dispatcher
│   ├── KeysActionFields.tsx
│   ├── ApiControlActionFields.tsx
│   ├── DirectControlActionFields.tsx
│   └── VirtualActionFields.tsx
└── assignments/
    ├── AssignmentsList.tsx
    ├── AssignmentCard.tsx
    ├── AssignmentHeader.tsx
    ├── AssignmentTypeSelector.tsx
    ├── EditConditionsModal.tsx
    └── {type}/
        └── {Type}Fields.tsx
```
