"use client";

import { Button } from "@/components/ui/button";
import { useApiMutation } from "@/lib/useApiMutation";

export function LeaveReviewButtons({ id }: { id: number }) {
  const review = useApiMutation<{ decision: "approved" | "rejected" }>(
    `/api/admin/leave-requests/${id}/review`,
  );

  const pending = review.isPending ? review.variables?.decision : null;

  return (
    <div className="flex items-center gap-2">
      <Button
        onClick={() => review.mutate({ decision: "approved" })}
        disabled={review.isPending}
        size="xs"
        className="bg-green-600 text-white hover:bg-green-500"
      >
        {pending === "approved" ? "…" : "Setujui"}
      </Button>
      <Button
        onClick={() => review.mutate({ decision: "rejected" })}
        disabled={review.isPending}
        size="xs"
        variant="destructive"
        className="bg-red-600 text-white hover:bg-red-500"
      >
        {pending === "rejected" ? "…" : "Tolak"}
      </Button>
      {review.isError && <span className="text-xs text-red-600">{review.error.message}</span>}
    </div>
  );
}
