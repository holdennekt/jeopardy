"use server";

import { ErrorDTO, isError } from "@/middleware";
import { PacksResp } from "./components/PacksList";
import { cookies } from "next/headers";
import { PackDTO } from "./components/pack/PackEditor";

const PAGE_QUERY_PARAM = "page";
const FILTER_QUERY_PARAM = "filter";

export const getPacks = async (packFilter: string, page?: number) => {
  const url = new URL(`http://${process.env.BACKEND_HOST}/rest/packs`);
  url.searchParams.set(FILTER_QUERY_PARAM, packFilter);
  if (page) url.searchParams.set(PAGE_QUERY_PARAM, page.toString());
  const resp = await fetch(url.toString(), {
    cache: "no-store",
    headers: { cookie: cookies().toString() },
  }).catch(console.log);
  const packs: PacksResp | ErrorDTO = await resp?.json();
  if (isError(packs)) throw new Error(packs.error);
  return packs;
};

export const getPack = async (id: string) => {
  const url = new URL(`http://${process.env.BACKEND_HOST}/rest/pack/${id}`);
  const resp = await fetch(url.toString(), {
    cache: "no-store",
    headers: { cookie: cookies().toString() },
  }).catch(console.log);
  const pack: PackDTO | ErrorDTO = await resp?.json();
  if (isError(pack)) throw new Error(pack.error);
  return pack;
};

export const createPack = async (pack: PackDTO) => {
  const url = new URL(`http://${process.env.BACKEND_HOST}/rest/pack`);
  const resp = await fetch(url.toString(), {
    method: "POST",
    headers: { cookie: cookies().toString() },
    body: JSON.stringify(pack),
  }).catch(console.log);
  const obj: { id: string } | ErrorDTO = await resp?.json();
  if (isError(obj)) throw new Error(obj.error);
  return obj;
};

export const updatePack = async (id: string, pack: PackDTO) => {
  const url = new URL(`http://${process.env.BACKEND_HOST}/rest/pack/${id}`);
  const resp = await fetch(url.toString(), {
    method: "PUT",
    headers: { cookie: cookies().toString() },
    body: JSON.stringify(pack),
  }).catch(console.log);
  const obj: { id: string } | ErrorDTO = await resp?.json();
  if (isError(obj)) throw new Error(obj.error);
  return obj;
};
