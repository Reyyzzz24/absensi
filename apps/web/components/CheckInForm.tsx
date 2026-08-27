"use client";

import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { useRouter } from "next/navigation";
import { Button } from "@/components/ui/button";
import { Checkbox } from "@/components/ui/checkbox";
import { Skeleton } from "@/components/ui/skeleton";
import { GeofenceMapLoader } from "@/components/geofence/GeofenceMapLoader";
import { distanceMeters } from "@/lib/geo";
import { cn } from "@/lib/utils";
import type { OfficeLocation } from "@/lib/types";
import { MapPinCheck, MapPinOff } from "lucide-react";

type Coords = { latitude: number; longitude: number };

type CameraState = "idle" | "starting" | "ready" | "denied" | "unavailable";
type LocationState = "idle" | "locating" | "ready" | "denied" | "unavailable";

export function CheckInForm({ geofences = [] }: { geofences?: OfficeLocation[] }) {
  const router = useRouter();
  const videoRef = useRef<HTMLVideoElement>(null);
  const streamRef = useRef<MediaStream | null>(null);

  const [cameraState, setCameraState] = useState<CameraState>("idle");
  const [locationState, setLocationState] = useState<LocationState>("idle");
  const [coords, setCoords] = useState<Coords | null>(null);
  const [photoDataUrl, setPhotoDataUrl] = useState<string | null>(null);
  const [isWfh, setIsWfh] = useState(false);
  const [submitting, setSubmitting] = useState(false);
  const [serverError, setServerError] = useState<string | null>(null);
  const [result, setResult] = useState<Record<string, unknown> | null>(null);

  const startCamera = useCallback(async () => {
    setCameraState("starting");
    try {
      const stream = await navigator.mediaDevices.getUserMedia({
        video: { facingMode: "user" },
        audio: false,
      });
      streamRef.current = stream;
      if (videoRef.current) {
        videoRef.current.srcObject = stream;
      }
      setCameraState("ready");
    } catch {
      setCameraState("denied");
    }
  }, []);

  const requestLocation = useCallback(() => {
    if (!("geolocation" in navigator)) {
      setLocationState("unavailable");
      return;
    }
    setLocationState("locating");
    navigator.geolocation.getCurrentPosition(
      (pos) => {
        setCoords({ latitude: pos.coords.latitude, longitude: pos.coords.longitude });
        setLocationState("ready");
      },
      () => setLocationState("denied"),
      { enableHighAccuracy: true, timeout: 10000 },
    );
  }, []);

  useEffect(() => {
    startCamera();
    requestLocation();
    return () => {
      streamRef.current?.getTracks().forEach((t) => t.stop());
    };
  }, [startCamera, requestLocation]);

  function capturePhoto() {
    const video = videoRef.current;
    if (!video) return;
    const canvas = document.createElement("canvas");
    canvas.width = video.videoWidth;
    canvas.height = video.videoHeight;
    const ctx = canvas.getContext("2d");
    if (!ctx) return;
    ctx.drawImage(video, 0, 0, canvas.width, canvas.height);
    setPhotoDataUrl(canvas.toDataURL("image/png"));
  }

  function retakePhoto() {
    setPhotoDataUrl(null);
  }

  async function submitCheckIn() {
    if (!coords || !photoDataUrl) return;
    setServerError(null);
    setSubmitting(true);
    try {
      // Backend expects raw base64, not a full data: URI.
      const base64 = photoDataUrl.split(",")[1] ?? "";

      const res = await fetch("/api/attendance/check-in", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          latitude: coords.latitude,
          longitude: coords.longitude,
          photo: base64,
          is_wfh: isWfh,
        }),
      });

      const data = await res.json().catch(() => ({}));
      if (!res.ok) {
        setServerError(mapError(res.status, data.error));
        return;
      }

      setResult(data);
      router.refresh();
    } catch {
      setServerError("Tidak bisa menghubungi server. Coba lagi.");
    } finally {
      setSubmitting(false);
    }
  }

  // Nearest active office location to the user's current coords -- used only
  // to draw the map/badge. Never affects canSubmit; the server (D-1) is the
  // sole authority on whether a check-in is actually within radius.
  //
  // Must run unconditionally, before the `if (result) return` below --
  // hooks can never come after an early return, since that makes the hook
  // count differ between the "just submitted" render and every other
  // render ("Rendered fewer hooks than expected").
  const nearest = useMemo(() => {
    if (!coords || geofences.length === 0) return null;
    let best = geofences[0];
    let bestDistance = distanceMeters(coords.latitude, coords.longitude, best.latitude, best.longitude);
    for (const loc of geofences.slice(1)) {
      const d = distanceMeters(coords.latitude, coords.longitude, loc.latitude, loc.longitude);
      if (d < bestDistance) {
        best = loc;
        bestDistance = d;
      }
    }
    return { location: best, distance: bestDistance };
  }, [coords, geofences]);

  if (result) {
    return (
      <div className="rounded-md bg-green-50 p-4 text-sm text-green-800">
        <p className="font-medium">Berhasil!</p>
        <p className="mt-1">Status presensi: {String(result.status)}</p>
        <Button onClick={() => router.push("/dashboard")} size="sm" className="mt-3">
          Kembali ke Dashboard
        </Button>
      </div>
    );
  }

  const canSubmit = cameraState === "ready" && locationState === "ready" && !!photoDataUrl && !submitting;
  const withinRadius = nearest ? nearest.distance <= nearest.location.radius_meters : null;

  return (
    <div className="space-y-4">
      <div className="overflow-hidden rounded-md bg-black">
        {!photoDataUrl && (
          <video ref={videoRef} autoPlay playsInline muted className="aspect-video w-full object-cover" />
        )}
        {photoDataUrl && (
          // eslint-disable-next-line @next/next/no-img-element
          <img src={photoDataUrl} alt="Foto check-in" className="aspect-video w-full object-cover" />
        )}
      </div>

      {cameraState === "denied" && (
        <p className="text-sm text-red-600">
          Akses kamera ditolak. Izinkan akses kamera di browser lalu muat ulang halaman.
        </p>
      )}
      {cameraState === "unavailable" && (
        <p className="text-sm text-red-600">Kamera tidak tersedia di perangkat ini.</p>
      )}

      <div className="flex gap-2">
        {cameraState === "ready" && !photoDataUrl && <Button onClick={capturePhoto}>Ambil Foto</Button>}
        {photoDataUrl && (
          <Button onClick={retakePhoto} variant="outline">
            Ambil Ulang
          </Button>
        )}
      </div>

      <div className="space-y-2 rounded-md bg-slate-50 p-3 text-sm">
        <p className="font-medium text-slate-700">Lokasi</p>
        {locationState === "locating" && (
          <>
            <p className="text-slate-500">Mendeteksi lokasi…</p>
            {geofences.length > 0 && <Skeleton className="h-40 w-full rounded-2xl" />}
          </>
        )}
        {locationState === "ready" && coords && (
          <div className="space-y-2">
            {nearest && (
              <GeofenceMapLoader
                mode="display"
                center={{ lat: nearest.location.latitude, lng: nearest.location.longitude }}
                radius={nearest.location.radius_meters}
                userPosition={{ lat: coords.latitude, lng: coords.longitude }}
                heightClassName="h-40"
              />
            )}
            <p className="text-slate-600">
              {coords.latitude.toFixed(6)}, {coords.longitude.toFixed(6)}
            </p>
            {nearest && !isWfh && (
              <span
                className={cn(
                  "inline-flex items-center gap-1.5 rounded-full px-2.5 py-1 text-xs font-medium",
                  withinRadius
                    ? "bg-[var(--status-hadir-bg)] text-[var(--status-hadir)]"
                    : "bg-[var(--status-alpha-bg)] text-[var(--status-alpha)]",
                )}
              >
                {withinRadius ? <MapPinCheck className="size-3.5" /> : <MapPinOff className="size-3.5" />}
                {withinRadius ? "Dalam radius" : "Di luar radius"} · {Math.round(nearest.distance)}m
              </span>
            )}
            {nearest && isWfh && (
              <span className="inline-flex items-center gap-1.5 rounded-full bg-blue-50 px-2.5 py-1 text-xs font-medium text-blue-700">
                <MapPinCheck className="size-3.5" />
                Mode WFH — radius kantor tidak berlaku
              </span>
            )}
          </div>
        )}
        {locationState === "denied" && (
          <p className="text-red-600">
            Akses lokasi ditolak. Izinkan akses lokasi di browser, lalu{" "}
            <button onClick={requestLocation} className="underline">
              coba lagi
            </button>
            .
          </p>
        )}
        {locationState === "unavailable" && (
          <p className="text-red-600">Geolocation tidak didukung di browser ini.</p>
        )}
      </div>

      <label className="flex items-center gap-2 text-sm text-slate-700">
        <Checkbox checked={isWfh} onCheckedChange={(checked) => setIsWfh(checked === true)} />
        Work From Home
      </label>

      {serverError && (
        <div className="rounded-md bg-red-50 px-3 py-2 text-sm text-red-700">{serverError}</div>
      )}

      <Button onClick={submitCheckIn} disabled={!canSubmit} className="w-full">
        {submitting ? "Memproses…" : "Kirim Presensi"}
      </Button>
    </div>
  );
}

function mapError(status: number, message?: string): string {
  if (status === 422) return "Lokasi Anda di luar radius kantor yang diizinkan.";
  if (status === 409) return message ?? "Terlalu cepat / sudah check-out untuk siklus ini.";
  return message ?? "Gagal mengirim presensi.";
}
