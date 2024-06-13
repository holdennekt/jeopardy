import Room, { RoomDTO } from "@/app/components/room/Room";
import { ErrorDTO, isError, UserDTO } from "@/middleware";
import { cookies } from "next/headers";

const PASSWORD_QUERY_PARAM = "password";

const connectToRoom = async (id: string, password: string | undefined) => {
  const url = new URL(`http://${process.env.BACKEND_HOST}/rest/room/${id}`);
  if (password) url.searchParams.set(PASSWORD_QUERY_PARAM, password);
  const resp = await fetch(url.toString(), {
    method: "PATCH",
    cache: "no-store",
    headers: { cookie: cookies().toString() },
  }).catch(console.log);
  const room: RoomDTO | ErrorDTO = await resp?.json();
  if (isError(room)) throw new Error(room.error);
}

export default async function Page({
  params,
  searchParams,
}: {
  params: { id: string };
  searchParams: { [key: string]: string | undefined };
}) {
  // const user: UserDTO = JSON.parse(headers().get(USER_HEADER_NAME)!);
  const user: UserDTO = { id: "1", name: "holdennekt", avatar: null };
  const room = await connectToRoom(params.id, searchParams.password);

  const currentQuestion = {
    index: 1,
    value: 100,
    text: "Where are pigeon babies?",
    attachment: null,
    answers: ["no one knows", "they do not exists", "they are being born as adults"],
    comment: "i mean seriously, have you ever seen one?",
  };

  // const currentQuestion = null;

  const dummyRoom: RoomDTO = {
    id: "123",
    name: "dungeon",
    packPreview: { id: "1", name: "test pack" },
    players: [
      { id: "1", name: "nikita", avatar: null, score: 0, isConnected: true },
      { id: "2", name: "julia", avatar: null, score: 0, isConnected: true },
      { id: "3", name: "danylo", avatar: null, score: 0, isConnected: true },
      { id: "4", name: "danya", avatar: null, score: 0, isConnected: true },
      { id: "5", name: "volodymyr", avatar: null, score: 0, isConnected: false },
      { id: "6", name: "sasha", avatar: null, score: 0, isConnected: true },
      { id: "7", name: "bohdan", avatar: null, score: 0, isConnected: true },
      { id: "8", name: "yana", avatar: null, score: 0, isConnected: true },
    ],
    host: { id: "9", name: "maksym", avatar: null, isConnected: true },
    currentRound: "first",
    availableQuestions: {
      category1: [
        { index: 0, value: 100, hasBeenPlayed: false },
        { index: 1, value: 200, hasBeenPlayed: false },
        { index: 2, value: 300, hasBeenPlayed: false },
        { index: 3, value: 400, hasBeenPlayed: true },
        { index: 4, value: 500, hasBeenPlayed: false },
      ],
      category2: [
        { index: 0, value: 100, hasBeenPlayed: false },
        { index: 1, value: 200, hasBeenPlayed: false },
        { index: 2, value: 300, hasBeenPlayed: false },
        { index: 3, value: 400, hasBeenPlayed: false },
        { index: 4, value: 500, hasBeenPlayed: true },
      ],
      category3: [
        { index: 0, value: 100, hasBeenPlayed: false },
        { index: 1, value: 200, hasBeenPlayed: true },
        { index: 2, value: 300, hasBeenPlayed: true },
        { index: 3, value: 400, hasBeenPlayed: false },
        { index: 4, value: 500, hasBeenPlayed: false },
      ],
      category4: [
        { index: 0, value: 100, hasBeenPlayed: false },
        { index: 1, value: 200, hasBeenPlayed: true },
        { index: 2, value: 300, hasBeenPlayed: false },
        { index: 3, value: 400, hasBeenPlayed: true },
        { index: 4, value: 500, hasBeenPlayed: false },
      ],
      category5: [
        { index: 0, value: 100, hasBeenPlayed: true },
        { index: 1, value: 200, hasBeenPlayed: false },
        { index: 2, value: 300, hasBeenPlayed: false },
        { index: 3, value: 400, hasBeenPlayed: false },
        { index: 4, value: 500, hasBeenPlayed: false },
      ],
    },
    currentPlayer: "1",
    currentQuestion,
    answeringPlayer: "1",
    allowedToAnswer: ["1", "2", "3", "4", "5", "6", "7", "8"],
    finalRoundState: {
      isActive: false,
      availableQuestions: null,
      question: null,
      bets: null,
    },
    isPaused: false,
  };

  return (
    <Room user={user} initialRoom={dummyRoom} />
  );
}
