import { Nav } from "@/_components/Nav";

export default function Home() {
  return (
    <>
      <Nav />

      <section className="h-full lg:grid lg:place-content-center">
        <div className="mx-auto w-screen max-w-7xl px-4 py-16 sm:px-6 sm:py-24 lg:px-8 lg:py-32">
          <div className="mx-auto max-w-prose text-center">
            <h1 className="font-title text-2xl md:text-3xl lg:text-4xl font-bold">
              TSW Controller App
            </h1>

            <p className="mt-4 text-base text-pretty text-base-content/80">
              The best way to control Train Sim World and Train Simulator Classic with any
              controller or joystick.
            </p>

            <div className="mt-4 flex justify-center gap-4 sm:mt-6">
              <a
                target="_blank"
                className="btn btn-primary"
                href="https://github.com/LiamMartens/tsw-controller-app/releases"
              >
                Download now
              </a>

              <a
                target="_blank"
                className="btn btn-outline"
                href="https://github.com/LiamMartens/tsw-controller-app"
              >
                Learn More
              </a>
            </div>
          </div>
        </div>
      </section>
    </>
  );
}
