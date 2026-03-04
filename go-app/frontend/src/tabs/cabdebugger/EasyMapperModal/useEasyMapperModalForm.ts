import { useForm, UseFormReturn } from "react-hook-form";
import { formSchema } from "./formSchema";
import { zodResolver } from "@hookform/resolvers/zod";
import z from "zod";

export type TUseEasyMapperModalFormReturn = UseFormReturn<z.infer<typeof formSchema>>

export const useEasyMapperModalForm = () => useForm<z.infer<typeof formSchema>>({
  resolver: zodResolver(formSchema),
  defaultValues: {
    controller: "",
    controls: [],
  },
});