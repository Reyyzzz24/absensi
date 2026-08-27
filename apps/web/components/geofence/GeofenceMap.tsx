"use client";

import { useEffect, useRef } from "react";
import L from "leaflet";
import "leaflet/dist/leaflet.css";

// Leaflet's default marker icon references image paths that don't survive
// bundling (the classic react-leaflet + webpack/Turbopack footgun -- markers
// render as broken image icons otherwise). Sidestepped entirely by using
// custom colored divIcons instead of L.Icon.Default, which also lets the
// office/user markers match the app's own design tokens instead of
// Leaflet's stock blue pin for everything.
function pinIcon(color: string) {
  return L.divIcon({
    className: "",
    html: `<svg width="30" height="42" viewBox="0 0 30 42" xmlns="http://www.w3.org/2000/svg">
      <path d="M15 0C6.7 0 0 6.7 0 15c0 10.5 15 27 15 27s15-16.5 15-27c0-8.3-6.7-15-15-15z" fill="${color}" stroke="white" stroke-width="1.5"/>
      <circle cx="15" cy="15" r="5.5" fill="white"/>
    </svg>`,
    iconSize: [30, 42],
    iconAnchor: [15, 42],
    popupAnchor: [0, -38],
  });
}

const officeIcon = pinIcon("#2F6BFF");
const userDotIcon = L.divIcon({
  className: "",
  html: `<div style="width:16px;height:16px;border-radius:9999px;background:#10B981;border:3px solid white;box-shadow:0 0 0 2px rgba(16,185,129,0.35),0 1px 3px rgba(0,0,0,0.3);"></div>`,
  iconSize: [16, 16],
  iconAnchor: [8, 8],
});

export type GeofenceMapProps = {
  mode: "display" | "edit";
  center: { lat: number; lng: number };
  radius: number;
  userPosition?: { lat: number; lng: number } | null;
  onChange?: (center: { lat: number; lng: number }) => void;
  draggable?: boolean;
  showRadius?: boolean;
  className?: string;
  heightClassName?: string;
};

const TILE_URL = process.env.NEXT_PUBLIC_MAP_TILE_URL || "https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png";

// Built directly on the Leaflet imperative API rather than react-leaflet's
// <MapContainer>. Next.js App Router's router.refresh() can re-render a
// client subtree via React's "reappear" path (used for Suspense/Activity-
// hidden trees) WITHOUT running the previous unmount cleanup first --
// react-leaflet's ref-callback then calls `new L.Map()` again on a DOM node
// Leaflet already marked as initialized, crashing with "Map container is
// already initialized" on every save. Managing the map instance ourselves
// in a single mount-effect (plus the _leaflet_id guard below) sidesteps the
// whole class of bug regardless of which React re-render path triggers it.
export function GeofenceMap({
  mode,
  center,
  radius,
  userPosition,
  onChange,
  draggable = mode === "edit",
  showRadius = true,
  className,
  heightClassName = "h-64",
}: GeofenceMapProps) {
  const containerRef = useRef<HTMLDivElement>(null);
  const mapRef = useRef<L.Map | null>(null);
  const officeMarkerRef = useRef<L.Marker | null>(null);
  const userMarkerRef = useRef<L.Marker | null>(null);
  const circleRef = useRef<L.Circle | null>(null);
  const onChangeRef = useRef(onChange);
  onChangeRef.current = onChange;

  useEffect(() => {
    const node = containerRef.current;
    if (!node) return;

    const stale = node as HTMLDivElement & { _leaflet_id?: number };
    if (stale._leaflet_id) delete stale._leaflet_id;

    const map = L.map(node, {
      center: [center.lat, center.lng],
      zoom: 16,
      scrollWheelZoom: mode === "edit",
    });
    mapRef.current = map;

    L.tileLayer(TILE_URL, {
      attribution: '&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors',
    }).addTo(map);

    if (showRadius) {
      circleRef.current = L.circle([center.lat, center.lng], {
        radius,
        color: "#2F6BFF",
        fillColor: "#2F6BFF",
        fillOpacity: 0.12,
        weight: 1.5,
      }).addTo(map);
    }

    const officeMarker = L.marker([center.lat, center.lng], { icon: officeIcon, draggable }).addTo(map);
    officeMarkerRef.current = officeMarker;
    if (draggable) {
      officeMarker.on("dragend", () => {
        const pos = officeMarker.getLatLng();
        onChangeRef.current?.({ lat: pos.lat, lng: pos.lng });
      });
    }

    if (mode === "edit") {
      map.on("click", (e: L.LeafletMouseEvent) => {
        onChangeRef.current?.({ lat: e.latlng.lat, lng: e.latlng.lng });
      });
    }

    if (userPosition) {
      userMarkerRef.current = L.marker([userPosition.lat, userPosition.lng], { icon: userDotIcon }).addTo(map);
    }

    return () => {
      map.remove();
      mapRef.current = null;
      officeMarkerRef.current = null;
      userMarkerRef.current = null;
      circleRef.current = null;
    };
    // Deliberately mount-once: subsequent prop changes are applied via the
    // imperative updates below rather than tearing down/recreating the map
    // (which is what caused the crash in the first place).
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  useEffect(() => {
    const map = mapRef.current;
    if (!map) return;
    map.setView([center.lat, center.lng], map.getZoom(), { animate: true });
    officeMarkerRef.current?.setLatLng([center.lat, center.lng]);
    circleRef.current?.setLatLng([center.lat, center.lng]);
  }, [center.lat, center.lng]);

  useEffect(() => {
    circleRef.current?.setRadius(radius);
  }, [radius]);

  useEffect(() => {
    const map = mapRef.current;
    if (!map) return;
    if (userPosition) {
      if (!userMarkerRef.current) {
        userMarkerRef.current = L.marker([userPosition.lat, userPosition.lng], { icon: userDotIcon }).addTo(map);
      } else {
        userMarkerRef.current.setLatLng([userPosition.lat, userPosition.lng]);
      }
    } else if (userMarkerRef.current) {
      userMarkerRef.current.remove();
      userMarkerRef.current = null;
    }
  }, [userPosition?.lat, userPosition?.lng]);

  return (
    <div
      ref={containerRef}
      className={`relative overflow-hidden rounded-2xl border border-border ${heightClassName} ${className ?? ""}`}
    />
  );
}
