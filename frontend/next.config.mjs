/** @type {import('next').NextConfig} */
const nextConfig = {
  reactStrictMode: false,
  async rewrites() {
    return [
      {
        source: '/api/:path*',
        destination: `http://${process.env.BACKEND_HOST}/:path*`,
      },
    ];
  },
};

export default nextConfig;
