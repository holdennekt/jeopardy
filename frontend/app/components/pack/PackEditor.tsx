"use client";

import React, { FormEventHandler, useState } from "react";
import { HiddenQuestion } from "../room/Room";
import { HiddenCategory } from "../PacksList";
import QuestionModal from "./QuestionModal";
import { toast, ToastContainer } from "react-toastify";
import { usePathname, useRouter } from "next/navigation";
import RoundEditor from "./RoundEditor";
import Accordion from "../Accordion";
import FinalCategoryModal from "./FinalCategoryModal";
import Link from "next/link";

export type PackDTO = {
  name: string;
  type: "public" | "private";
  rounds: Round[];
  finalRound: FinalRound;
};
export type Round = {
  name: string;
  categories: Category[];
};
export type Category = HiddenCategory & {
  questions: Question[];
};
export type Answer = {
  answers: string[];
  comment: string | null;
};
export type Question = HiddenQuestion & Answer;
export type FinalRound = {
  categories: FinalCategory[];
};
export type FinalCategory = {
  name: string;
  question: FinalQuestion;
};
export type FinalQuestion = Answer & {
  text: string;
  attachment: {
    mediaType: "image" | "audio" | "video";
    contentUrl: string;
  } | null;
};

export default function PackEditor({
  handlePack,
  initialPack,
  readOnly = false,
}: {
  handlePack: (pack: PackDTO) => Promise<{ id: string }>;
  initialPack?: PackDTO;
  readOnly?: boolean;
}) {
  const router = useRouter();
  const pathname = usePathname();

  const [pack, setPack] = useState<PackDTO>(
    initialPack ?? {
      name: "",
      type: "public",
      rounds: [{ name: "", categories: [] }],
      finalRound: { categories: [] },
    }
  );
  const [finalCategoryNameInput, setFinalCategoryNameInput] = useState("");
  const [questionModal, setQuestionModal] = useState<{
    isOpen: boolean;
    roundIndex: number;
    categoryIndex: number;
    questionIndex: number;
    question: Question;
  }>({
    isOpen: false,
    roundIndex: -1,
    categoryIndex: -1,
    questionIndex: -1,
    question: {
      index: 0,
      value: 0,
      text: "",
      attachment: null,
      answers: [],
      comment: null,
    },
  });
  const [finalCategoryModal, setFinalCategoryModal] = useState<{
    isOpen: boolean;
    index: number;
    category: FinalCategory;
  }>({
    isOpen: false,
    index: -1,
    category: {
      name: "",
      question: {
        text: "",
        attachment: null,
        answers: [],
        comment: null,
      },
    },
  });

  const addRound = () => {
    pack.rounds.push({ name: "", categories: [] });
    setPack({ ...pack });
  };

  const changeQuestion = (
    roundIndex: number,
    categoryIndex: number,
    questionIndex: number,
    question: Question
  ) => {
    pack.rounds[roundIndex].categories[categoryIndex].questions[questionIndex] =
      { ...question, index: questionModal.questionIndex };
    setPack({ ...pack });
  };

  const addFinalCategory = (name: string) => {
    pack.finalRound.categories.push({
      name,
      question: { text: "", attachment: null, answers: [], comment: null },
    });
    setPack({ ...pack });
  };

  const changeFinalCategory = (index: number, category: FinalCategory) => {
    pack.finalRound.categories[index] = category;
    setPack({ ...pack });
  };

  const onSubmit: FormEventHandler<HTMLFormElement> = async (e) => {
    e.preventDefault();

    try {
      const obj = await handlePack(pack);
      const url = `/packs/${obj.id}`;
      router.push(url);
      toast.success("Pack successfully saved!", { containerId: "editor" });
    } catch (error) {
      if (error instanceof Error)
        return toast.error(error.message, { containerId: "editor" });
    }
  };

  return (
    <>
      <form className="min-h-0 h-full flex flex-col gap-4" onSubmit={onSubmit}>
        <div className="flex items-end gap-4">
          <label>
            <p className="font-medium">Pack name</p>
            <input
              className="w-48 h-8 rounded-md mt-1 p-1 text-black"
              type="text"
              placeholder="Name"
              value={pack.name}
              onChange={(e) => setPack({ ...pack, name: e.target.value })}
              readOnly={readOnly}
            />
          </label>
          <label>
            <p className="font-medium">Privacy Type</p>
            {readOnly ? (
              <input
                className="w-48 h-8 rounded-md mt-1 p-1 text-black"
                type="text"
                value={pack.type}
                readOnly={readOnly}
              />
            ) : (
              <select
                className="w-48 h-8 mt-1 p-0.5 rounded-md text-black"
                value={pack.type}
                onChange={(e) =>
                  setPack({
                    ...pack,
                    type: e.target.value as "public" | "private",
                  })
                }
              >
                <option value="public">Public</option>
                <option value="private">Private</option>
              </select>
            )}
          </label>
          {!readOnly && (
            <button
              className="w-fit h-fit rounded px-2 py-1 primary"
              type="button"
              onClick={addRound}
            >
              Add round
            </button>
          )}
        </div>
        <div className="flex-1 flex flex-col gap-2 overflow-y-auto">
          {pack.rounds.map((round, roundIndex) => (
            <Accordion title={`Round ${roundIndex + 1}`} key={roundIndex}>
              <RoundEditor
                round={round}
                index={roundIndex}
                pack={pack}
                setPack={setPack}
                setQuestionModal={setQuestionModal}
                readOnly={readOnly}
              />
            </Accordion>
          ))}
          <Accordion title="Final round">
            {!readOnly && (
              <>
                <label>
                  <p className="mt-2 font-medium">Category name</p>
                  <input
                    className="w-48 h-8 rounded-md mt-1 p-1 text-black"
                    type="text"
                    placeholder="Name"
                    value={finalCategoryNameInput}
                    onChange={(e) => setFinalCategoryNameInput(e.target.value)}
                    readOnly={readOnly}
                  />
                </label>
                <button
                  className="w-fit h-fit ml-4 rounded px-2 py-1 primary"
                  type="button"
                  onClick={() => {
                    addFinalCategory(finalCategoryNameInput);
                    setFinalCategoryNameInput("");
                  }}
                >
                  Add category
                </button>
              </>
            )}
            <ul className="flex flex-col gap-2 mt-4 pl-4 list-inside list-disc">
              {pack.finalRound.categories.map((category, index) => (
                <li key={index}>
                  <button
                    className="w-fit h-fit px-2 py-1 border rounded"
                    type="button"
                    onClick={() =>
                      setFinalCategoryModal({ isOpen: true, index, category })
                    }
                  >
                    {category.name}
                  </button>
                </li>
              ))}
            </ul>
          </Accordion>
        </div>
        {readOnly ? (
          <div className="flex flex-row-reverse">
            <Link
              className="w-fit h-fit rounded px-2 py-1 primary"
              href={`${usePathname()}?edit=true`}
            >
              Edit
            </Link>
          </div>
        ) : (
          <div className="flex flex-row-reverse">
            <div>
              <button
                className="w-fit h-fit px-2 py-1 border rounded"
                type="button"
                onClick={() => {
                  router.push(pathname.split("?")[0]);
                }}
              >
                Discard
              </button>
              <button className="w-fit h-fit ml-4 px-2 py-1 rounded primary">
                Save
              </button>
            </div>
          </div>
        )}
      </form>
      <QuestionModal
        isOpen={questionModal.isOpen}
        close={() => setQuestionModal({ ...questionModal, isOpen: false })}
        question={questionModal.question}
        saveQuestion={changeQuestion.bind(
          null,
          questionModal.roundIndex,
          questionModal.categoryIndex,
          questionModal.questionIndex
        )}
        readOnly={readOnly}
      />
      <FinalCategoryModal
        isOpen={finalCategoryModal.isOpen}
        close={() =>
          setFinalCategoryModal({ ...finalCategoryModal, isOpen: false })
        }
        category={finalCategoryModal.category}
        saveCategory={changeFinalCategory.bind(null, finalCategoryModal.index)}
        readOnly={readOnly}
      />
      <ToastContainer
        containerId="editor"
        position="bottom-left"
        theme="colored"
      />
    </>
  );
}
