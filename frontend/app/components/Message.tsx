import Image from "next/image";
import { UserDTO } from "../../middleware";

export type ChatMessageDTO = {
  from: UserDTO;
  text: string;
};

const dummyChatMessage: ChatMessageDTO = {
  from: { id: "1", name: "user", avatar: null },
  text: "hello",
};

export const isChatMessageDTO = (obj: unknown): obj is ChatMessageDTO => {
  if (typeof obj !== 'object' || obj === null) return false;
  return Object.keys(dummyChatMessage).every((key) => Object.hasOwn(obj, key));
};

export default function Message({
  message,
  isOwn,
  isPrevUserSame,
  isNextUserSame,
}: {
  message: ChatMessageDTO;
  isOwn: boolean;
  isPrevUserSame: boolean;
  isNextUserSame: boolean;
}) {
  const emptyAvatar = <div className="h-8 aspect-square"></div>;
  const imgAvatar = (
    <Image
      className="h-8 aspect-square rounded-full"
      src={message.from.avatar!}
      alt='avatar'
    />
  );
  const divAvatar = (
    <div className="flex justify-center items-center h-8 aspect-square rounded-full bg-indigo-500">
      {message.from.name
        .split(" ")
        .map(word => word[0].toUpperCase())
        .join("")}
    </div>
  );

  return (
    <div
      className={
        `flex items-center gap-2 ${isPrevUserSame ? "mt-0.5" : "mt-1.5"}` +
        (isOwn ? " flex-row-reverse" : "")
      }
    >
      {!isOwn &&
        (isNextUserSame ?
          emptyAvatar :
          message.from.avatar ?
            imgAvatar :
            divAvatar)}
      <div
        className={`relative text-sm break-words rounded-xl py-2 px-4 ${
          isOwn ? "secondary" : "background"
        }`}
      >
        {!(isOwn || isPrevUserSame) && (
          <p className="username">{message.from.name}</p>
        )}
        <p>{message.text}</p>
      </div>
    </div>
  );
}
