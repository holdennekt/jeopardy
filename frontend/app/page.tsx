import Navbar from "./components/Navbar";
import { cookies, headers } from "next/headers";
import Lobby from "./components/lobby/Lobby";
import { ErrorDTO, isError, USER_HEADER_NAME, UserDTO } from "../middleware";
import { LobbyRoomDTO } from "./components/lobby/LobbyRoom";

const getRooms = async () => {
  const url = new URL(`http://${process.env.BACKEND_HOST}/rest/rooms`);
  const resp = await fetch(url, {
    cache: "no-store",
    headers: { cookie: cookies().toString() },
  });
  const rooms: LobbyRoomDTO[] | ErrorDTO = await resp?.json();
  if (isError(rooms)) throw new Error(rooms.error);
  return rooms;
};

export default async function Home() {
  const user: UserDTO = JSON.parse(headers().get(USER_HEADER_NAME)!);
  const rooms = await getRooms();

  return (
    <>
      <Navbar user={user} />
      <Lobby user={user} initialRooms={rooms} />
    </>
  );
}
