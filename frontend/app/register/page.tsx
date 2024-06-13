"use client";

import { toast, ToastContainer } from "react-toastify";
import Navbar from "../components/Navbar";
import { FormEventHandler } from "react";
import { ErrorDTO, isError } from "@/middleware";
import { useRouter } from "next/navigation";
import Link from "next/link";

export default function Page() {
  const router = useRouter();
  const onSubmit: FormEventHandler<HTMLFormElement> = async (e) => {
    e.preventDefault();

    const resp = await fetch(
      "/api/register",
      {
        method: "POST",
        body: new FormData(e.currentTarget),
      },
    ).catch(console.log);
    const obj: { id: string; } | ErrorDTO = await resp?.json();

    if (isError(obj)) return toast.error(obj.error, { containerId: "register" });
    router.push("/");
  };

  return (
    <>
      <Navbar />
      <main className="flex-1 flex justify-center items-center">
        <div className="rounded-xl p-5 surface">
          <form action="" onSubmit={onSubmit}>
            <label className="block">
              <p className="text-sm font-medium">Login</p>
              <input
                className="w-full h-8 rounded-lg mt-1 p-1 text-black"
                type="text"
                placeholder="Login"
                name="login"
                minLength={1}
                maxLength={50}
                required
              />
            </label>
            <label className="block mt-2">
              <p className="text-sm font-medium">Password</p>
              <input
                className="w-full h-8 rounded-lg mt-1 p-1 text-black"
                type="password"
                placeholder="Password"
                name="password"
                minLength={1}
                maxLength={50}
                required
              />
            </label>
            <div className="flex justify-between mt-3">
              <Link
                className="rounded px-3 py-1 font-medium secondary"
                href={"/login"}
              >
                Sign in
              </Link>
              <button className="rounded px-3 py-1 font-medium primary">
                Sign up
              </button>
            </div>
          </form>
        </div>
      </main>
      <ToastContainer
        containerId="register"
        position="bottom-left"
        theme="colored"
      />
    </>
  );
}
