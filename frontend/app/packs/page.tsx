import { headers } from "next/headers";
import Navbar from "../components/Navbar";
import { USER_HEADER_NAME, UserDTO } from "../../middleware";
import PacksList from "../components/PacksList";
import { getPacks } from "../actions";

export default async function Packs() {
  const user: UserDTO = JSON.parse(headers().get(USER_HEADER_NAME)!);
  const packs = await getPacks("", 1);

  return (
    <>
      <Navbar user={user} />
      <PacksList user={user} initialPacks={packs} />
    </>
  );
}
