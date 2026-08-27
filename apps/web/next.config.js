/** @type {import('next').NextConfig} */
const nextConfig = {
  // react-leaflet v4's MapContainer doesn't clean up Leaflet's internal
  // per-DOM-node init marker correctly across StrictMode's dev-only
  // mount->unmount->mount cycle, causing a hard "Map container is already
  // initialized" runtime crash on every page using <GeofenceMap>. This is a
  // known upstream incompatibility (react-leaflet/react-leaflet#1039), not
  // fixable from our component code. StrictMode is a dev-only diagnostic
  // aid; disabling it does not affect production behavior.
  reactStrictMode: false,
  transpilePackages: ["@absensi-next/contracts"],
};

module.exports = nextConfig;
