import type {
  AdminUser,
  DisasterEvent,
  EventListResponse,
  LoginResponseWritable,
  SessionResponse,
} from "@/lib/api/generated/types.gen"

export type { AdminUser, DisasterEvent, EventListResponse }

export type AdminSession = SessionResponse &
  Partial<Pick<LoginResponseWritable, "csrfToken">>
