import { zodResolver } from "@hookform/resolvers/zod"
import { useMutation, useQueryClient } from "@tanstack/react-query"
import { CircleAlertIcon, LockKeyholeIcon } from "lucide-react"
import { useForm } from "react-hook-form"
import { z } from "zod"

import { AtlasMark } from "@/components/atlas-mark"
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert"
import { Button } from "@/components/ui/button"
import {
  Field,
  FieldDescription,
  FieldError,
  FieldGroup,
  FieldLabel,
} from "@/components/ui/field"
import { Input } from "@/components/ui/input"
import { Spinner } from "@/components/ui/spinner"
import { adminSessionQuery, createAdminSession } from "@/features/admin/api"
import { APIError } from "@/lib/api/client"

z.config({ jitless: true })

const loginSchema = z.object({
  email: z.email("Enter a valid email address"),
  password: z.string().min(12, "Password must be at least 12 characters"),
})

type LoginValues = z.infer<typeof loginSchema>

export function LoginForm() {
  const queryClient = useQueryClient()
  const form = useForm<LoginValues>({
    resolver: zodResolver(loginSchema),
    defaultValues: { email: "", password: "" },
  })
  const loginMutation = useMutation({
    mutationFn: createAdminSession,
    onSuccess: (session) => {
      queryClient.setQueryData(adminSessionQuery.queryKey, session)
    },
  })

  return (
    <main className="grid min-h-svh place-items-center bg-muted/45 px-4 py-12">
      <section
        aria-labelledby="login-title"
        className="w-full max-w-sm rounded-xl border bg-background p-7 shadow-sm"
      >
        <div className="flex items-center gap-3">
          <AtlasMark />
          <div>
            <p className="font-semibold">Atlas</p>
            <p className="text-xs text-muted-foreground">
              Administration portal
            </p>
          </div>
        </div>
        <div className="mt-8">
          <h1
            id="login-title"
            className="text-2xl font-semibold tracking-tight"
          >
            Sign in to Atlas
          </h1>
          <p className="mt-2 text-sm leading-6 text-muted-foreground">
            Review evidence and publish verified incident updates.
          </p>
        </div>

        <form
          className="mt-7"
          onSubmit={form.handleSubmit((values) => loginMutation.mutate(values))}
          noValidate
        >
          <FieldGroup>
            <Field data-invalid={Boolean(form.formState.errors.email)}>
              <FieldLabel htmlFor="email">Email address</FieldLabel>
              <Input
                id="email"
                type="email"
                autoComplete="username"
                aria-invalid={Boolean(form.formState.errors.email)}
                {...form.register("email")}
              />
              <FieldError>{form.formState.errors.email?.message}</FieldError>
            </Field>
            <Field data-invalid={Boolean(form.formState.errors.password)}>
              <FieldLabel htmlFor="password">Password</FieldLabel>
              <Input
                id="password"
                type="password"
                autoComplete="current-password"
                aria-invalid={Boolean(form.formState.errors.password)}
                {...form.register("password")}
              />
              <FieldError>{form.formState.errors.password?.message}</FieldError>
            </Field>
            {loginMutation.isError && (
              <Alert variant="destructive">
                <CircleAlertIcon aria-hidden="true" />
                <AlertTitle>Sign-in failed</AlertTitle>
                <AlertDescription>
                  {loginMutation.error instanceof APIError
                    ? loginMutation.error.message
                    : "Atlas could not reach the identity service."}
                </AlertDescription>
              </Alert>
            )}
            <Button
              type="submit"
              size="lg"
              disabled={loginMutation.isPending}
              className="w-full"
            >
              {loginMutation.isPending ? (
                <Spinner data-icon="inline-start" aria-hidden="true" />
              ) : (
                <LockKeyholeIcon data-icon="inline-start" aria-hidden="true" />
              )}
              Sign in
            </Button>
            <FieldDescription className="text-center">
              Access is limited to authorized Atlas editors.
            </FieldDescription>
          </FieldGroup>
        </form>
      </section>
    </main>
  )
}
