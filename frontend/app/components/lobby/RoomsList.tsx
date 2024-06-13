import LobbyRoom, { LobbyRoomDTO } from "./LobbyRoom";

export default function RoomsList({
  rooms,
  openPasswordModal,
}: {
  rooms: LobbyRoomDTO[];
  openPasswordModal: (roomId: string) => void;
}) {
  return (
    <div className={`flex flex-col gap-2 flex-auto min-w-0 min-h-0 mt-3 
      overflow-x-clip overflow-y-auto rounded surface`}>
      {rooms.length ? rooms.map((room, index) => (
        <LobbyRoom
          key={index}
          room={room}
          openPasswordModal={openPasswordModal}
        />
      )) : <p className="text-center">No rooms yet</p>}
    </div>
  );
}
